package application

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/CZMA-eng/urlshortener/config"
	"github.com/CZMA-eng/urlshortener/database"
	"github.com/CZMA-eng/urlshortener/internal/api"
	"github.com/CZMA-eng/urlshortener/internal/cache"
	"github.com/CZMA-eng/urlshortener/internal/service"
	"github.com/CZMA-eng/urlshortener/pkg/shortcode"
	"github.com/CZMA-eng/urlshortener/pkg/validator"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)



type Application struct {
	e *echo.Echo
	db *sql.DB
	redisClient *cache.RedisCache
	urlService 	*service.URLService
	urlHandler *api.URLHandler
	cfg *config.Config
	shortCodeGenerator *shortcode.ShortCode
}

func (a *Application) Init(filePath string) error{
	cfg, err := config.LoadConfig(filePath)
	if err != nil {
		return fmt.Errorf("load error: %w", err)
	}

	a.cfg = cfg
	db, err := database.NewDB(cfg.Database)
	if err != nil {
		return err
	}

	a.db = db

	redisClient, err := cache.NewRedisCache(cfg.Redis)
	if err != nil {
		return err
	}
	a.redisClient = redisClient

	a.shortCodeGenerator = shortcode.NewShortCode(cfg.ShortCode.Length)

	a.urlService = service.NewURLService(db, a.shortCodeGenerator, cfg.App.DefaultDuration, redisClient, cfg.App.BaseURL)

	a.urlHandler = api.NewURLHandler(a.urlService)

	e := echo.New()
	e.Server.WriteTimeout = cfg.Server.WriteTimeout
	e.Server.ReadTimeout = cfg.Server.ReadTimeout
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	
	e.POST("/api/url", a.urlHandler.CreateURL)
	e.GET("/:code", a.urlHandler.RedirectURL)
	e.Validator = validator.NewCustomValidator()
	a.e = e
	return nil
}
 
func (a * Application) Run(){
	go a.startServer()
	go a.cleanUp()
	go a.shutdown()
}

func (a * Application) startServer() {
	if err := a.e.Start(a.cfg.Server.Addr); err != nil{
		log.Println(err)
	}
}

func (a *Application) cleanUp(){
	ticker := time.NewTicker(a.cfg.App.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		if err := a.urlService.DeleteURL(context.Background()); err != nil{
			log.Println(err)
		}
	}
}

func (a *Application) shutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	defer func(){
		if err := a.db.Close(); err != nil{
			log.Println(err)
		}
	}()

	defer func() {
		if err := a.redisClient.Close(); err != nil{
			log.Println(err)
		}
	}()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := a.e.Shutdown(ctx);err != nil{
			log.Println(err)
		}

}