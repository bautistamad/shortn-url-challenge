package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"url-shortener/internal/domain/entities"
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
		expectedLocation string
		expectedStatus   int
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

				assert.Equal(t, tc.expectedLocation, location.String())
			}
		})
	}
}

func getShortenURLEndpointTestCases() []struct {
	mockService      ports.MockShorternServiceType
	requestBody      interface{}
	name             string
	expectedShortURL string
	expectedStatus   int
	expectError      bool
} {
	return []struct {
		mockService      ports.MockShorternServiceType
		requestBody      interface{}
		name             string
		expectedShortURL string
		expectedStatus   int
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

func TestHandleGetUrlStats(t *testing.T) {
	tests := []struct {
		mockService    ports.MockShorternServiceType
		name           string
		shortURL       string
		expectedBody   string
		expectedStatus int
	}{
		{
			name:     "success",
			shortURL: "abc123",
			mockService: ports.MockShorternServiceType{
				GetUrlStats: &ports.UrlStatsOrError{
					Url: &entities.URL{
						ShortURL:    "abc123",
						LongURL:     "http://example.com",
						AccessCount: 10,
					},
					Err: nil,
				},
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"longURL":"http://example.com","shortUrl":"abc123","id":"00000000-0000-0000-0000-000000000000","accessCount":10}`,
		},
		{
			name:     "failed",
			shortURL: "notfound",
			mockService: ports.MockShorternServiceType{
				GetUrlStats: &ports.UrlStatsOrError{
					Err: ports.ErrUrlNotFound,
				},
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "URL not found",
		},
		{
			name:     "server failed",
			shortURL: "error",
			mockService: ports.MockShorternServiceType{
				GetUrlStats: &ports.UrlStatsOrError{
					Err: errors.New("internal server error"),
				},
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Internal server error: internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := ports.NewMockShorternService(tt.mockService)

			handler := &HTTPHandler{
				shortenerService: mockService,
			}

			req, err := http.NewRequest("GET", "/stats/"+tt.shortURL, nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()

			router := mux.NewRouter()
			router.HandleFunc("/stats/{shortURL}", handler.handleGetUrlStats).Methods("GET")

			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				var response entities.URL
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(t, err)

				expectedResponse := entities.URL{
					ShortURL:    "abc123",
					LongURL:     "http://example.com",
					AccessCount: 10,
				}

				assert.Equal(t, expectedResponse.ShortURL, response.ShortURL)
				assert.Equal(t, expectedResponse.LongURL, response.LongURL)
				assert.Equal(t, expectedResponse.AccessCount, response.AccessCount)
			} else {
				assert.Equal(t, tt.expectedBody, rr.Body.String())
			}
		})
	}
}

func TestHandleDeleteURL(t *testing.T) {
	tests := []struct {
		mockService    ports.MockShorternServiceType
		name           string
		shortURL       string
		expectedBody   string
		expectedStatus int
	}{
		{
			name:     "success",
			shortURL: "abc123",
			mockService: ports.MockShorternServiceType{
				DeleteUrl: &ports.DeleteUrlOrError{
					DeletedUrl: "http://example.com",
					Err:        nil,
				},
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"URL deleted successfully","deleted_url":"http://example.com"}`,
		},
		{
			name:     "failed not found",
			shortURL: "notfound",
			mockService: ports.MockShorternServiceType{
				DeleteUrl: &ports.DeleteUrlOrError{
					Err: ports.ErrUrlNotFound,
				},
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "URL not found",
		},
		{
			name:     "failed server error",
			shortURL: "error",
			mockService: ports.MockShorternServiceType{
				DeleteUrl: &ports.DeleteUrlOrError{
					Err: errors.New("internal server error"),
				},
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Internal server error: internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := ports.NewMockShorternService(tt.mockService)

			handler := &HTTPHandler{
				shortenerService: mockService,
			}

			req, err := http.NewRequest("DELETE", "/url/"+tt.shortURL, nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()

			router := mux.NewRouter()
			router.HandleFunc("/url/{shortURL}", handler.handleDeleteURL).Methods("DELETE")

			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			actualBody := strings.TrimSpace(rr.Body.String())

			assert.Equal(t, tt.expectedBody, actualBody)
		})
	}
}
