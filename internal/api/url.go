package api

import (
	"context"
	"net/http"

	"github.com/CZMA-eng/urlshortener/internal/model"
	"github.com/labstack/echo/v4"
)

type URLService interface{
	CreateURL(ctx context.Context, req model.CreateURLRequest)(*model.
		CreateURLResponse, error)

	GetURL(ctx context.Context, shortcode string) (string, error)
}

type URLHandler struct{
	urlService URLService
}

func NewURLHandler(urlService URLService) *URLHandler {
	return &URLHandler{
		urlService: urlService,
	}
}

// POST /api/url original_url, custome_code, duration, => short_url, expire time
func (h * URLHandler) CreateURL(c echo.Context) error {
	// extract data
	var req model.CreateURLRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error()) 
	}

	// validate data type
	if err := c.Validate(req);err != nil{
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// call function
	resp, err := h.urlService.CreateURL(c.Request().Context(), req)
	if err != nil{
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	//give response
	return c.JSON(http.StatusCreated, resp)
}

// GET /: code redirect short url to long-url
func (h * URLHandler) RedirectURL(c echo.Context) error {
	// get code 
	shortCode := c.Param("code")

	// shortcode => url
	originalURL, err := h.urlService.GetURL(c.Request().Context(), shortCode)
	if err != nil{
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.Redirect(http.StatusPermanentRedirect, originalURL)
}