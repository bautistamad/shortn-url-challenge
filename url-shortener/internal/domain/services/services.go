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
	errSavingUrl = errors.New("failed to save url")
	errDeleteUrl = errors.New("failed to delete url")
	errGetUrl    = errors.New("failed to get long url")
)

type Shortener struct {
	cacheRepository ports.UrlRepository
	dbRepository    ports.UrlRepository
}

var _ ports.ShorternService = new(Shortener)

func NewShortenerService(cacheRepo ports.UrlRepository, dbRepo ports.UrlRepository) *Shortener {
	return &Shortener{
		cacheRepository: cacheRepo,
		dbRepository:    dbRepo,
	}
}

func (s *Shortener) CreateShortUrl(longUrl string) (string, error) {

	shortUrl, err := s.dbRepository.SaveShortenUrl(defaultShortUrl+generateShortKey(), longUrl)

	if err != nil {
		err = errors.Join(errSavingUrl, err)
		return "", err
	}

	return shortUrl, nil
}

func (s *Shortener) DeleteUrl(shortUrl string) (string, error) {

	url := entities.URL{
		ShortURL: shortUrl,
	}

	deletedUrl, err := s.dbRepository.DeleteShortenUrl(url)

	if err != nil {
		err = errors.Join(errDeleteUrl, err)
		return "", err
	}

	_, err = s.cacheRepository.GetLongUrl(shortUrl)

	if err != nil {
		log.Printf("url is not in cache %s", deletedUrl)
		return deletedUrl, nil
	}
	_, err = s.cacheRepository.DeleteShortenUrl(url)

	if err != nil {
		log.Println("error deleted data from cache")
	}

	return deletedUrl, nil
}

func (s *Shortener) GetLongUrl(key string) (string, error) {

	longUrl, err := s.cacheRepository.GetLongUrl(defaultShortUrl + key)

	if err != nil {
		log.Printf("url not found on cache %s", longUrl)
	}

	longUrl, err = s.dbRepository.GetLongUrl(defaultShortUrl + key)

	if err != nil {
		return "", errors.Join(errGetUrl, err)
	}

	_, err = s.cacheRepository.SaveShortenUrl(defaultShortUrl+key, longUrl)

	if err != nil {
		log.Printf("cant save url on cache %s", defaultShortUrl+key)
	}

	return longUrl, nil
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
