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
	errUpdateURL      = errors.New("failed to update url in postgres")
)

type PostgresRepository struct {
	client *gorm.DB
}

var _ ports.DatabaseUrlRepository = new(PostgresRepository)

func GetPGClient(config PostgresConfig) (ports.DatabaseUrlRepository, error) {
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
func (d *PostgresRepository) DeleteShortenUrl(url entities.URL) error {
	result := d.client.Where("shorturl = ?", url.ShortURL).Delete(url)
	if result.Error != nil {
		return errors.Join(errDeleteURL, result.Error)
	}

	if result.RowsAffected == 0 {
		return ports.ErrUrlNotFound
	}

	return nil
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

	return resultUrl.LongURL, result.Error
}

// func (d *PostgresRepository) GetLongUrlByLongUrl(longUrl string) (string, error) {
// 	var resultUrl entities.URL

// 	result := d.client.Where("longurl = ?", longUrl).Select("longurl").First(&resultUrl)

// 	if result.Error != nil {
// 		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
// 			return "", ports.ErrUrlNotFound
// 		}

// 		return "", errGetURL
// 	}

// 	return resultUrl.LongURL, result.Error
// }

func (d *PostgresRepository) GetUrlStats(shortUrl string) (*entities.URL, error) {
	var url entities.URL
	result := d.client.Where("shorturl = ?", shortUrl).First(&url)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return &entities.URL{}, ports.ErrUrlNotFound
		}
		return &entities.URL{}, errors.Join(errGetURL, result.Error)
	}
	return &url, nil
}

func (d *PostgresRepository) IncrementAccessCount(shortUrl string) error {
	result := d.client.Model(&entities.URL{}).Where("shorturl = ?", shortUrl).
		Update("access_count", gorm.Expr("access_count + 1"))
	if result.Error != nil {
		return errors.Join(errUpdateURL, result.Error)
	}
	return nil
}
