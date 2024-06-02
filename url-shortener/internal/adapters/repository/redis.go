// internal/adapters/repository/redis_repository.go
package repository

import (
	"context"
	"errors"
	"url-shortener/internal/domain/entities"
	"url-shortener/internal/domain/ports"

	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Address  string
	Password string
	Port     string
	DB       int
}

var (
	ErrCreateClient = errors.New("failed to create Redis client")
	ErrSavingUrl    = errors.New("failed to save url in redis")
	ErrGettingUrl   = errors.New("failed to get url from redis")
	ErrDeletingUrl  = errors.New("failed to delete url in redis")
)

type RedisRepository struct {
	client *redis.Client
}

var _ ports.UrlRepository = new(RedisRepository)

func GetRedisClient(config RedisConfig) (ports.UrlRepository, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     config.Address + ":" + config.Port,
		Password: config.Password,
		DB:       config.DB,
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, ErrCreateClient
	}

	return &RedisRepository{client: client}, nil
}

func (d *RedisRepository) GetLongUrl(shortUrl string) (string, error) {
	ctx := context.Background()

	longUrl, err := d.client.Get(ctx, shortUrl).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil
		}
		return "", ErrGettingUrl
	}

	return longUrl, nil
}

func (d *RedisRepository) SaveShortenUrl(shortUrl string, longUrl string) (string, error) {
	ctx := context.Background()

	err := d.client.Set(ctx, shortUrl, longUrl, 0).Err()
	if err != nil {
		return "", ErrSavingUrl
	}
	return shortUrl, nil
}

func (d *RedisRepository) DeleteShortenUrl(url entities.URL) (string, error) {
	ctx := context.Background()

	err := d.client.Del(ctx, url.ShortURL).Err()
	if err != nil {
		return "", ErrDeletingUrl
	}

	return url.ShortURL, nil
}
