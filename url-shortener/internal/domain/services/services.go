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

	shortUrlKey := defaultShortUrl + generateShortKey()
	shortUrl, err := s.dbRepository.SaveShortenUrl(shortUrlKey, longUrl)

	if err != nil {
		err = errors.Join(errSavingUrl, err)
		return "", err
	}

	_, err = s.cacheRepository.SaveShortenUrl(shortUrl, longUrl)

	if err != nil {
		log.Println("cant save data on cache")
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
		if !errors.Is(err, ports.ErrUrlNotFound) {
			log.Printf("error getting URL from cache: %v", err)
		}

	} else if longUrl != "" {
		if err = s.dbRepository.IncrementAccessCount(defaultShortUrl + key); err != nil {
			log.Printf("error incrementing access count for URL %s: %v", key, err)
		}
		return longUrl, nil
	}

	longUrl, err = s.dbRepository.GetLongUrl(defaultShortUrl + key)

	if err != nil {
		if errors.Is(err, ports.ErrUrlNotFound) {
			log.Printf("url not found on database")
			return "", nil
		}
		log.Printf("error getting URL from database: %v", err)
		return "", errGetUrl
	}

	if _, err := s.cacheRepository.SaveShortenUrl(defaultShortUrl+key, longUrl); err != nil {
		if !errors.Is(err, ports.ErrUrlNotFound) {
			log.Printf("can't save URL on cache: %v", err)
		}
	}

	if err := s.dbRepository.IncrementAccessCount(defaultShortUrl + key); err != nil {
		log.Printf("error incrementing access count for URL %s: %v", key, err)
	}

	return longUrl, nil
}

func (s *Shortener) GetUrlStats(key string) (*entities.URL, error) {
	shortUrl := defaultShortUrl + key

	urlStats, err := s.dbRepository.GetUrlStats(shortUrl)
	if err != nil {
		return &entities.URL{}, errors.Join(errGetUrlStats, err)
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
