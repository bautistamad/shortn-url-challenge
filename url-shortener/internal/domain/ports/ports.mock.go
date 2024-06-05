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

	MockDatabaseUrlRepository struct {
		mock.Mock
	}

	MockDatabaseUrlRepositoryType struct {
		GetLongUrl           *LongUrlOrErr
		CreateShortUrl       *CreateShortUrlOrErr
		DeleteShortenUrl     *ErrOnlyRet
		GetUrlStats          *UrlStatsOrError
		IncrementAccessCount *ErrOnlyRet
	}

	MockCacheUrlRepository struct {
		mock.Mock
	}

	MockCacheUrlRepositoryType struct {
		GetLongUrl       *LongUrlOrErr
		SaveShortenUrl   *CreateShortUrlOrErr
		DeleteShortenUrl *ErrOnlyRet
	}

	LongUrlOrErr struct {
		Err     error
		LongURL string
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

	ErrOnlyRet struct {
		Err error
	}
)

var (
	_ ShorternService       = new(MockShorternService)
	_ DatabaseUrlRepository = new(MockDatabaseUrlRepository)
	_ CacheUrlRepository    = new(MockCacheUrlRepository)
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
		service.On("GetLongUrl", mock.AnythingOfType("string")).Return(mkSer.GetLongUrl.LongURL, mkSer.GetLongUrl.Err)
	}

	if mkSer.GetUrlStats != nil {
		service.On("GetUrlStats", mock.AnythingOfType("string")).Return(mkSer.GetUrlStats.Url, mkSer.GetUrlStats.Err)
	}

	return service
}

func NewMockDatabaseUrlRepository(mkRepo MockDatabaseUrlRepositoryType) *MockDatabaseUrlRepository {
	repo := &MockDatabaseUrlRepository{}

	if mkRepo.CreateShortUrl != nil {
		repo.On("SaveShortenUrl", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(mkRepo.CreateShortUrl.ShortUrl, mkRepo.CreateShortUrl.Err)
	}

	if mkRepo.DeleteShortenUrl != nil {
		repo.On("DeleteShortenUrl", mock.AnythingOfType("entities.URL")).Return(mkRepo.DeleteShortenUrl.Err)
	}

	if mkRepo.GetLongUrl != nil {
		repo.On("GetLongUrl", mock.AnythingOfType("string")).Return(mkRepo.GetLongUrl.LongURL, mkRepo.GetLongUrl.Err)
	}

	if mkRepo.GetUrlStats != nil {
		repo.On("GetUrlStats", mock.AnythingOfType("string")).Return(mkRepo.GetUrlStats.Url, mkRepo.GetUrlStats.Err)
	}

	if mkRepo.IncrementAccessCount != nil {
		repo.On("IncrementAccessCount", mock.AnythingOfType("string")).Return(mkRepo.IncrementAccessCount.Err)
	}

	return repo
}

func NewMockCacheUrlRepository(mkRepo MockCacheUrlRepositoryType) *MockCacheUrlRepository {
	repo := &MockCacheUrlRepository{}

	if mkRepo.SaveShortenUrl != nil {
		repo.On("SaveShortenUrl", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(mkRepo.SaveShortenUrl.ShortUrl, mkRepo.SaveShortenUrl.Err)
	}

	if mkRepo.DeleteShortenUrl != nil {
		repo.On("DeleteShortenUrl", mock.AnythingOfType("entities.URL")).Return(mkRepo.DeleteShortenUrl.Err)
	}

	if mkRepo.GetLongUrl != nil {
		repo.On("GetLongUrl", mock.AnythingOfType("string")).Return(mkRepo.GetLongUrl.LongURL, mkRepo.GetLongUrl.Err)
	}

	return repo
}

func (m *MockShorternService) DeleteUrl(shortUrl string) (string, error) {
	args := m.Called(shortUrl)
	return args.String(0), args.Error(1)
}

func (m *MockShorternService) GetLongUrl(key string) (string, error) {
	args := m.Called(key)
	return args.String(0), args.Error(1)
}

func (m *MockShorternService) GetUrlStats(shortUrl string) (*entities.URL, error) {
	args := m.Called(shortUrl)

	if args.Get(0).(*entities.URL) == nil {

		return nil, args.Error(1)

	}
	return args.Get(0).(*entities.URL), args.Error(1)
}

func (m *MockShorternService) CreateShortUrl(longUrl string) (string, error) {
	args := m.Called(longUrl)
	return args.String(0), args.Error(1)
}

func (m *MockDatabaseUrlRepository) DeleteShortenUrl(url entities.URL) error {
	args := m.Called(url)
	return args.Error(0)
}

func (m *MockDatabaseUrlRepository) GetLongUrl(key string) (string, error) {
	args := m.Called(key)
	return args.String(0), args.Error(1)
}

func (m *MockDatabaseUrlRepository) GetUrlStats(shortUrl string) (*entities.URL, error) {
	args := m.Called(shortUrl)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.URL), args.Error(1)
}

func (m *MockDatabaseUrlRepository) IncrementAccessCount(shortUrl string) error {
	args := m.Called(shortUrl)
	return args.Error(0)
}

func (m *MockDatabaseUrlRepository) SaveShortenUrl(shortUrl, longUrl string) (string, error) {
	args := m.Called(shortUrl, longUrl)
	return args.String(0), args.Error(1)
}

func (m *MockCacheUrlRepository) DeleteShortenUrl(url entities.URL) error {
	args := m.Called(url)
	return args.Error(0)
}

func (m *MockCacheUrlRepository) GetLongUrl(key string) (string, error) {
	args := m.Called(key)
	return args.String(0), args.Error(1)
}

func (m *MockCacheUrlRepository) SaveShortenUrl(shortUrl, longUrl string) (string, error) {
	args := m.Called(shortUrl, longUrl)
	return args.String(0), args.Error(1)
}
