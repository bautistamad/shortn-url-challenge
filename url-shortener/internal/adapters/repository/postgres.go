package repository

import (
	"errors"
	"fmt"
	"time"
	"url-shortener/internal/domain/entities"
	"url-shortener/internal/domain/ports"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostgresConfig struct {
	Host                  string
	User                  string
	Password              string
	DBName                string
	ApplicationName       string
	Port                  int
	MaxOpenConnections    int
	MaxIdleConnections    int
	ConnectionMaxLifeTime time.Duration
	SSL                   bool
}

var (
	errCreatePGClient = errors.New("failed to create postgres client")
	errSaveURL        = errors.New("failed save new url to postgres")
	errDeleteURL      = errors.New("failed to delete url in postgres")
	errGetURL         = errors.New("failed to get url from postgres")
)

type PostgresRepository struct {
	client *gorm.DB
}

var _ ports.UrlRepository = new(PostgresRepository)

func GetPGClient(config PostgresConfig) (ports.UrlRepository, error) {
	connectionString := getConnectionString(config)

	client, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  connectionString,
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		QueryFields: true,
	})

	if err != nil {
		return nil, errors.Join(errCreatePGClient, err)
	}

	return &PostgresRepository{client: client}, nil

}

func getConnectionString(config PostgresConfig) string {
	var sslMode string
	if config.SSL {
		sslMode = "enable"
	} else {
		sslMode = "disable"
	}

	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d application_name=%s sslmode=%s",
		config.Host,
		config.User,
		config.Password,
		config.DBName,
		config.Port,
		config.ApplicationName,
		sslMode)
}

// SaveShortenUrl implements ports.DBRepository.
func (d *PostgresRepository) SaveShortenUrl(shortUrl string, longUrl string) (string, error) {

	var newUrl entities.URL

	newUrl.LongURL = longUrl
	newUrl.ShortURL = shortUrl
	res := d.client.Create(&newUrl)

	if res.Error != nil {
		return "", errors.Join(errSaveURL, res.Error)
	}

	return newUrl.ShortURL, nil
}

// DeleteShortenUrl implements ports.DBRepository.
func (d *PostgresRepository) DeleteShortenUrl(url entities.URL) (string, error) {
	result := d.client.Where("shorturl = ?", url.ShortURL).Delete(url)

	if result.Error != nil {
		return "", errors.Join(errDeleteURL, result.Error)
	}

	return url.ShortURL, nil
}

func (d *PostgresRepository) GetLongUrl(shortUrl string) (string, error) {

	var resultUrl entities.URL

	result := d.client.Where("shorturl = ?", shortUrl).Select("longurl").First(&resultUrl)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return "", ports.ErrUrlNotFound
		}

		return "", errGetURL
	}

	return resultUrl.LongURL, nil
}
