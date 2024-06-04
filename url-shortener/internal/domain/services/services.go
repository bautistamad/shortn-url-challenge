package services

import (
	"errors"
	"log"
	"math/rand"
	"time"
	"url-shortener/internal/domain/entities"
	"url-shortener/internal/domain/ports"
)

const (
	defaultShortUrl = "http://localhost/"
)

var (
	errSavingUrl   = errors.New("failed to save url")
	errDeleteUrl   = errors.New("failed to delete url")
	errGetUrl      = errors.New("failed to get long url")
	errGetUrlStats = errors.New("failed to get url stats")
)

type Shortener struct {
	cacheRepository ports.CacheUrlRepository
	dbRepository    ports.DatabaseUrlRepository
}

var _ ports.ShorternService = new(Shortener)

func NewShortenerService(cacheRepo ports.CacheUrlRepository, dbRepo ports.DatabaseUrlRepository) *Shortener {
	return &Shortener{
		cacheRepository: cacheRepo,
		dbRepository:    dbRepo,
	}
}

func (s *Shortener) CreateShortUrl(longUrl string) (string, error) {
	// validate if long url exist in db
	shortUrl, err := s.dbRepository.SaveShortenUrl(defaultShortUrl+generateShortKey(), longUrl)

	if err != nil {
		err = errors.Join(errSavingUrl, err)
		return "", err
	}

	return shortUrl, nil
}

func (s *Shortener) DeleteUrl(key string) (string, error) {
	url := entities.URL{
		ShortURL: defaultShortUrl + key,
	}

	err := s.dbRepository.DeleteShortenUrl(url)
	if err != nil {
		if errors.Is(err, ports.ErrUrlNotFound) {
			return "", ports.ErrUrlNotFound
		}
		return "", errors.Join(errDeleteUrl, err)
	}

	urlToDelete, err := s.cacheRepository.GetLongUrl(url.ShortURL)
	if err != nil {
		if errors.Is(err, ports.ErrUrlNotFound) {
			log.Printf("url is not in cache %s", url.ShortURL)
			return "", nil
		}
	} else {
		err = s.cacheRepository.DeleteShortenUrl(url)
		if err != nil {
			log.Println("error deleting data from cache")
		}
	}

	return urlToDelete, nil
}

func (s *Shortener) GetLongUrl(key string) (string, error) {
	longUrl, err := s.cacheRepository.GetLongUrl(defaultShortUrl + key)

	if err != nil {
		return "", errors.Join(errGetUrl, err)
	}

	if longUrl == "" {
		log.Printf("url not found on cache %s", longUrl)

		longUrl, err = s.dbRepository.GetLongUrl(defaultShortUrl + key)
		if err != nil {
			return "", errors.Join(errGetUrl, err)
		}

		_, err = s.cacheRepository.SaveShortenUrl(defaultShortUrl+key, longUrl)
		if err != nil {
			log.Printf("cant save url on cache %s", defaultShortUrl+key)
		}
	}

	err = s.dbRepository.IncrementAccessCount(key)

	if err != nil {
		log.Printf("error incrementing access count for url %s: %v", key, err)
	}

	return longUrl, nil
}

func (s *Shortener) GetUrlStats(key string) (entities.URL, error) {
	shortUrl := defaultShortUrl + key

	// Obtener las estadísticas de la base de datos
	urlStats, err := s.dbRepository.GetUrlStats(shortUrl)
	if err != nil {
		return entities.URL{}, errors.Join(errGetUrlStats, err)
	}

	return urlStats, nil
}

func generateShortKey() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const keyLength = 6

	rand.New(rand.NewSource(time.Now().UnixNano()))
	shortKey := make([]byte, keyLength)
	for i := range shortKey {
		shortKey[i] = charset[rand.Intn(len(charset))]
	}
	return string(shortKey)
}
