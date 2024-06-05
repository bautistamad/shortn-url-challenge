package services

import (
	"testing"
	"url-shortener/internal/domain/entities"
	"url-shortener/internal/domain/ports"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateShortUrl(t *testing.T) {
	t.Parallel()

	for _, tc := range getCreateShortUrlTestCases() {
		tc := tc // Necesario para evitar que se sobrescriban los valores en el bucle
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockDbRepo := ports.NewMockDatabaseUrlRepository(tc.dbMockRepo)
			mockCacheRepo := ports.NewMockCacheUrlRepository(tc.cacheMockRepo)

			service := NewShortenerService(mockCacheRepo, mockDbRepo)

			result, err := service.CreateShortUrl(tc.longUrl)

			if tc.expectedError {
				require.Error(t, err)
				assert.Empty(t, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.dbMockRepo.CreateShortUrl.ShortUrl, result)
			}
		})
	}
}

type (
	getCreateUrlTestCase struct {
		dbMockRepo    ports.MockDatabaseUrlRepositoryType
		cacheMockRepo ports.MockCacheUrlRepositoryType
		name          string
		longUrl       string
		expectedError bool
	}
)

func getCreateShortUrlTestCases() []getCreateUrlTestCase {
	return []getCreateUrlTestCase{
		{
			name:    "success",
			longUrl: "https://example.com/123/321/3456/23131",
			dbMockRepo: ports.MockDatabaseUrlRepositoryType{
				CreateShortUrl: &ports.CreateShortUrlOrErr{
					ShortUrl: "http://short.url/abc123",
					Err:      nil,
				},
			},
			cacheMockRepo: ports.MockCacheUrlRepositoryType{
				SaveShortenUrl: &ports.CreateShortUrlOrErr{
					ShortUrl: "http://short.url/abc123",
					Err:      nil,
				},
			},
			expectedError: false,
		},
		{
			name:    "repository error",
			longUrl: "https://example.com/123/321/3456/23131",
			dbMockRepo: ports.MockDatabaseUrlRepositoryType{
				CreateShortUrl: &ports.CreateShortUrlOrErr{
					ShortUrl: "",
					Err:      errSavingUrl,
				},
			},
			cacheMockRepo: ports.MockCacheUrlRepositoryType{
				SaveShortenUrl: &ports.CreateShortUrlOrErr{},
			},
			expectedError: true,
		},
	}
}

func TestGetUrlStats(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		dbMockRepo     ports.MockDatabaseUrlRepositoryType
		expectedError  error
		expectedResult *entities.URL
		name           string
		key            string
	}{
		{
			name: "success",
			key:  "abc123",
			dbMockRepo: ports.MockDatabaseUrlRepositoryType{
				GetUrlStats: &ports.UrlStatsOrError{
					Err: nil,
					Url: &entities.URL{
						ShortURL: "http://localhost/abc123",
						LongURL:  "http://example.com",
						// other fields...
					},
				},
			},
			expectedResult: &entities.URL{
				ShortURL: "http://localhost/abc123",
				LongURL:  "http://example.com",
			},
			expectedError: nil,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockDbRepo := ports.NewMockDatabaseUrlRepository(tc.dbMockRepo)
			service := NewShortenerService(nil, mockDbRepo)

			result, err := service.GetUrlStats(tc.key)

			assert.Equal(t, tc.expectedResult, result)
			assert.Equal(t, tc.expectedError, err)
		})
	}
}

type (
	getUrlTestCase struct {
		expectedError bool
		dbMockRepo    ports.MockDatabaseUrlRepositoryType
		cacheMockRepo ports.MockCacheUrlRepositoryType
		name          string
		shortUrl      string
	}
)

func TestGetUrl(t *testing.T) {
	t.Parallel()

	for _, tc := range getUrlCases() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockDbRepo := ports.NewMockDatabaseUrlRepository(tc.dbMockRepo)
			mockCacheRepo := ports.NewMockCacheUrlRepository(tc.cacheMockRepo)

			service := NewShortenerService(mockCacheRepo, mockDbRepo)

			result, err := service.GetLongUrl(tc.shortUrl)

			if tc.expectedError {
				require.Error(t, err)
				assert.Empty(t, result)
			}
			if tc.expectedError {
				require.NoError(t, err)
				assert.Equal(t, tc.cacheMockRepo.GetLongUrl.LongURL, result)
			}

			if tc.expectedError {
				require.NoError(t, err)
				assert.Equal(t, tc.dbMockRepo.GetLongUrl.LongURL, result)
			}

		})
	}
}

func getUrlCases() []getUrlTestCase {

	return []getUrlTestCase{
		{
			name:     "success cache",
			shortUrl: "http://localhost/123",
			cacheMockRepo: ports.MockCacheUrlRepositoryType{
				GetLongUrl: &ports.LongUrlOrErr{
					LongURL: "http://localhost/123",
					Err:     nil,
				},
			},
			dbMockRepo: ports.MockDatabaseUrlRepositoryType{
				GetLongUrl:           &ports.LongUrlOrErr{},
				IncrementAccessCount: &ports.ErrOnlyRet{},
			},
			expectedError: false,
		},
		{
			name:     "success db",
			shortUrl: "http://localhost/123",
			cacheMockRepo: ports.MockCacheUrlRepositoryType{
				GetLongUrl: &ports.LongUrlOrErr{
					Err:     nil,
					LongURL: "",
				},
				SaveShortenUrl: &ports.CreateShortUrlOrErr{},
			},
			dbMockRepo: ports.MockDatabaseUrlRepositoryType{
				GetLongUrl: &ports.LongUrlOrErr{
					LongURL: "http://localhost/123",
					Err:     nil,
				},
				IncrementAccessCount: &ports.ErrOnlyRet{},
			},
			expectedError: false,
		},
	}
}

type (
	deleteUrlTestCase struct {
		dbMockRepo    ports.MockDatabaseUrlRepositoryType
		cacheMockRepo ports.MockCacheUrlRepositoryType
		name          string
		key           string
		expectedError bool
	}
)

func getDeleteUrlCases() []deleteUrlTestCase {
	return []deleteUrlTestCase{
		{
			name: "success",
			key:  "123",
			dbMockRepo: ports.MockDatabaseUrlRepositoryType{
				DeleteShortenUrl: &ports.ErrOnlyRet{
					Err: nil,
				},
			},
			cacheMockRepo: ports.MockCacheUrlRepositoryType{
				DeleteShortenUrl: &ports.ErrOnlyRet{
					Err: nil,
				},
				GetLongUrl: &ports.LongUrlOrErr{
					Err:     nil,
					LongURL: "http://test.com/123",
				},
			},
			expectedError: false,
		},
		{
			name: "failed to delete",
			key:  "123",
			dbMockRepo: ports.MockDatabaseUrlRepositoryType{
				DeleteShortenUrl: &ports.ErrOnlyRet{
					Err: errDeleteUrl,
				},
			},
			cacheMockRepo: ports.MockCacheUrlRepositoryType{
				DeleteShortenUrl: &ports.ErrOnlyRet{},
				GetLongUrl:       &ports.LongUrlOrErr{},
			},
			expectedError: true,
		},
	}
}

func TestDeletetUrl(t *testing.T) {
	t.Parallel()

	for _, tc := range getDeleteUrlCases() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockDbRepo := ports.NewMockDatabaseUrlRepository(tc.dbMockRepo)
			mockCacheRepo := ports.NewMockCacheUrlRepository(tc.cacheMockRepo)

			service := NewShortenerService(mockCacheRepo, mockDbRepo)

			result, err := service.DeleteUrl("123")

			if tc.expectedError {
				require.Error(t, err)
				assert.Empty(t, result)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
