package ports

import (
	"errors"
	"url-shortener/internal/domain/entities"
)

var (
	ErrUrlNotFound = errors.New("failed top get shorturl")
)

type HTTPShortenHandler interface {
}

type UrlRepository interface {
	GetLongUrl(string) (string, error)
	SaveShortenUrl(string, string) (string, error)
	DeleteShortenUrl(entities.URL) (string, error)
}

type ShorternService interface {
	GetLongUrl(string) (string, error)
	CreateShortUrl(string) (string, error)
	DeleteUrl(string) (string, error)
}
