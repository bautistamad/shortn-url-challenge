package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"url-shortener/internal/domain/ports"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type (
	ShortenServiceTestCase struct {
		mockService      ports.MockShorternServiceType
		name             string
		path             string
		expectedStatus   int
		expectedLocation string
	}
)

func getShortenServiceTestCases() []ShortenServiceTestCase {
	return []ShortenServiceTestCase{
		{
			name: "success - return short url",
			mockService: ports.MockShorternServiceType{
				GetLongUrl: &ports.LongUrlOrErr{
					LongURL: "http://test.com/test123test123",
					Err:     nil,
				},
			},
			path:             "/short",
			expectedStatus:   http.StatusMovedPermanently,
			expectedLocation: "http://test.com/test123test123",
		},
		{
			name: "not found",
			mockService: ports.MockShorternServiceType{
				GetLongUrl: &ports.LongUrlOrErr{
					Err: ports.ErrUrlNotFound,
				},
			},
			path:           "/notfound",
			expectedStatus: http.StatusNotFound,
		},
	}
}

func TestRedirectEndpoint(t *testing.T) {
	for _, tc := range getShortenServiceTestCases() {
		t.Run(tc.name, func(t *testing.T) {
			router := mux.NewRouter()
			mockService := ports.NewMockShorternService(tc.mockService)
			handler := NewHTTPHandler(router, mockService)
			handler.RegisterRoutes()

			req, err := http.NewRequestWithContext(
				context.Background(),
				http.MethodGet,
				"http://localhost:8080"+tc.path,
				nil,
			)
			require.NoError(t, err)

			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			assert.Equal(t, tc.expectedStatus, recorder.Code)

			if tc.expectedStatus == http.StatusMovedPermanently {
				location, err := recorder.Result().Location()
				require.NoError(t, err)

				// Verificar que la ubicación de la redirección es la esperada
				assert.Equal(t, tc.expectedLocation, location.String())
			}
		})
	}
}

func getShortenURLEndpointTestCases() []struct {
	name             string
	requestBody      interface{}
	mockService      ports.MockShorternServiceType
	expectedStatus   int
	expectedShortURL string
	expectError      bool
} {
	return []struct {
		name             string
		requestBody      interface{}
		mockService      ports.MockShorternServiceType
		expectedStatus   int
		expectedShortURL string
		expectError      bool
	}{
		{
			name: "success",
			requestBody: map[string]string{
				"long_url": "http://example.com",
			},
			mockService: ports.MockShorternServiceType{
				CreateShortUrl: &ports.CreateShortUrlOrErr{
					ShortUrl: "http://short.url/abc123",
					Err:      nil,
				},
			},
			expectedStatus:   http.StatusOK,
			expectedShortURL: "http://short.url/abc123",
			expectError:      false,
		},
		{
			name:           "bad request",
			requestBody:    "invalid request body",
			mockService:    ports.MockShorternServiceType{},
			expectedStatus: http.StatusBadRequest,
			expectError:    false,
		},
		{
			name: "internal server error",
			requestBody: map[string]string{
				"long_url": "http://example.com",
			},
			mockService: ports.MockShorternServiceType{
				CreateShortUrl: &ports.CreateShortUrlOrErr{
					Err: errors.New("internal server error"),
				},
			},
			expectedStatus: http.StatusInternalServerError,
			expectError:    false,
		},
	}
}

func TestShortenURLEndpoint(t *testing.T) {
	for _, tc := range getShortenURLEndpointTestCases() {
		t.Run(tc.name, func(t *testing.T) {
			router := mux.NewRouter()
			mockService := ports.NewMockShorternService(tc.mockService)
			handler := NewHTTPHandler(router, mockService)
			handler.RegisterRoutes()

			requestBody, err := json.Marshal(tc.requestBody)
			require.NoError(t, err)

			req, err := http.NewRequestWithContext(
				context.Background(),
				http.MethodPost,
				"/shorten",
				bytes.NewReader(requestBody),
			)
			require.NoError(t, err)

			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			assert.Equal(t, tc.expectedStatus, recorder.Code)

			if tc.expectedStatus == http.StatusOK {
				var response struct {
					ShortURL string `json:"short_url"`
				}
				err := json.NewDecoder(recorder.Body).Decode(&response)
				require.NoError(t, err)
				assert.Equal(t, tc.expectedShortURL, response.ShortURL)
			}
		})
	}
}
