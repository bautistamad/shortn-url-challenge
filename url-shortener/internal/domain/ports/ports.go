package ports

import (
	"errors"
	"url-shortener/internal/domain/entities"
)

var (
	ErrUrlNotFound = errors.New("failed to get shorturl")
)

type CacheUrlRepository interface {
	GetLongUrl(string) (string, error)
	SaveShortenUrl(string, string) (string, error)
	DeleteShortenUrl(entities.URL) error
}

type DatabaseUrlRepository interface {
	GetLongUrl(string) (string, error)
	SaveShortenUrl(string, string) (string, error)
	DeleteShortenUrl(entities.URL) error
	GetLongUrlByLongUrl(string) (string, error)
	GetUrlStats(string) (*entities.URL, error)
	IncrementAccessCount(shortUrl string) error
}

type ShorternService interface {
	GetLongUrl(string) (string, error)
	CreateShortUrl(string) (string, error)
	DeleteUrl(string) (string, error)
	GetUrlStats(string) (*entities.URL, error)
}
