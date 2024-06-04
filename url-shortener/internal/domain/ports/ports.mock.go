package ports

import (
	"errors"
	"url-shortener/internal/domain/entities"

	"github.com/stretchr/testify/mock"
)

var ErrRandom = errors.New("random error")

type (
	MockShorternService struct {
		mock.Mock
	}

	MockShorternServiceType struct {
		GetLongUrl     *LongUrlOrErr
		CreateShortUrl *CreateShortUrlOrErr
		DeleteUrl      *DeleteUrlOrError
		GetUrlStats    *UrlStatsOrError
	}

	LongUrlOrErr struct {
		Err     error
		LongUrl string
	}

	CreateShortUrlOrErr struct {
		Err      error
		ShortUrl string
	}

	DeleteUrlOrError struct {
		Err        error
		DeletedUrl string
	}

	UrlStatsOrError struct {
		Err error
		Url *entities.URL
	}
)

var (
	_ ShorternService = new(MockShorternService)
)

func NewMockShorternService(mkSer MockShorternServiceType) *MockShorternService {
	service := &MockShorternService{}

	if mkSer.CreateShortUrl != nil {
		service.On("CreateShortUrl", mock.AnythingOfType("string")).Return(mkSer.CreateShortUrl.ShortUrl, mkSer.CreateShortUrl.Err)
	}

	if mkSer.DeleteUrl != nil {
		service.On("DeleteUrl", mock.AnythingOfType("string")).Return(mkSer.DeleteUrl.DeletedUrl, mkSer.DeleteUrl.Err)
	}

	if mkSer.GetLongUrl != nil {
		service.On("GetLongUrl", mock.AnythingOfType("string")).Return(mkSer.GetLongUrl.LongUrl, mkSer.GetLongUrl.Err)
	}

	if mkSer.GetUrlStats != nil {
		service.On("GetUrlStats", mock.AnythingOfType("string")).Return(mkSer.GetUrlStats.Url, mkSer.GetUrlStats.Err)
	}

	return service
}

// DeleteUrl implements ShorternService.
func (m *MockShorternService) DeleteUrl(shortUrl string) (string, error) {
	args := m.Called(shortUrl)
	return args.String(0), args.Error(1)
}

// GetLongUrl implements ShorternService.
func (m *MockShorternService) GetLongUrl(key string) (string, error) {
	args := m.Called(key)
	return args.String(0), args.Error(1)
}

// GetUrlStats implements ShorternService.
func (m *MockShorternService) GetUrlStats(shortUrl string) (*entities.URL, error) {
	args := m.Called(shortUrl)

	if args.Get(0).(*entities.URL) == nil {

		return nil, args.Error(1)

	}
	return args.Get(0).(*entities.URL), args.Error(1)
}

// CreateShortUrl implements ShorternService.
func (m *MockShorternService) CreateShortUrl(longUrl string) (string, error) {
	args := m.Called(longUrl)
	return args.String(0), args.Error(1)
}
