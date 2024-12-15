package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/CZMA-eng/urlshortener/internal/model"
	"github.com/CZMA-eng/urlshortener/internal/repo"
	"github.com/pingcap/errors"
)

type ShortCodeGenerator interface{
	GenerateShortCode() string
}

type Cacher interface {
	SetURL(ctx context.Context, url repo.Url) error
	GetURL(ctx context.Context, shortCode string)(*repo.Url, error)
}

type URLService struct {
	querier repo.Querier
	shortCodeGenerator ShortCodeGenerator
	defaultDuration time.Duration
	cache Cacher
	baseURL string
}

func NewURLService(db *sql.DB, shortCodeGenerator ShortCodeGenerator, duration time.Duration, cache Cacher, baseURL string) *URLService {
	return &URLService{
		querier:repo.New(db),
		shortCodeGenerator:shortCodeGenerator ,
		defaultDuration: duration,
		cache: cache,
		baseURL: baseURL,
	}
}

func (s *URLService) CreateURL(ctx context.Context, req model.CreateURLRequest)(*model.CreateURLResponse, error){
	var shortCode string
	var isCustom bool
	var expiredAt time.Time

	if req.CustomCode != ""{
		isAvailable, err := s.querier.IsShortCodeAvailable(ctx, req.CustomCode)
		if err != nil{
			return nil, err
		}
		if !isAvailable {
			return nil, fmt.Errorf("alias already exists")
		}
		shortCode = req.CustomCode
		isCustom=true
	}else{
		code, err := s.getShortCode(ctx, 0)
		if err != nil{
			return nil, err
		}
		shortCode = code
	}

	if req.Duration == nil {
		expiredAt = time.Now().Add(s.defaultDuration)
	}else{
		expiredAt = time.Now().Add(time.Hour * time.Duration(*req.Duration))
	}
	

	// insert into db
	url ,err := s.querier.CreateURL(ctx, repo.CreateURLParams{
		OriginalUrl: req.OriginalURL,
		ShortCode: shortCode,
		IsCustom: isCustom,
		ExpiredAt: expiredAt,
	})
	if err != nil{
		return nil, err
	}

	// store in redis
	if err := s.cache.SetURL(ctx, url); err != nil{
		return nil, err
	}

	return &model.CreateURLResponse{
		ShortURL: s.baseURL+"/"+url.ShortCode,
		ExpiredAT: url.ExpiredAt,
	}, nil
}

func (s *URLService) GetURL(ctx context.Context, shortCode string) (string, error){
	// visit cache
	url, err := s.cache.GetURL(ctx, shortCode)
	if err != nil{
		return "", err
	}

	if url != nil {
		return url.OriginalUrl, nil
	}

	// visit database
	url2, err := s.querier.GetUrlByShortCode(ctx, shortCode)
	if err != nil{
		return "", err
	}

	// store in cache
	if err := s.cache.SetURL(ctx, url2); err != nil{
		return "", err
	}

	return url2.OriginalUrl, nil
}

func (s *URLService) getShortCode(ctx context.Context, n int) (string, error){
	if n > 5 {
		return "", errors.New("too many attempts")
	}

	shortCode  := s.shortCodeGenerator.GenerateShortCode()

	isAvailable, err := s.querier.IsShortCodeAvailable(ctx, shortCode)
	if err != nil{
		return "", err
	}

	if isAvailable{
		return shortCode, nil
	}

	return s.getShortCode(ctx, n+1)
}

func (s *URLService) DeleteURL(ctx context.Context) error {
	return s.querier.DeleteURLExpired(ctx)
}