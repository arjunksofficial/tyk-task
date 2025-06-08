package ratelimit_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/arjunksofficial/tyk-task/internal/middlewares/ratelimit"
	"github.com/arjunksofficial/tyk-task/internal/token/models"
	"github.com/arjunksofficial/tyk-task/internal/token/services"
	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRateLimitMiddleware_RateLimitHandler(t *testing.T) {
	currentTime := time.Now()
	validTime := currentTime.Add(24 * time.Hour)
	validTimeStamp := validTime.UTC().Format(time.RFC3339)

	testCases := []struct {
		desc           string
		mockTokenSvc   services.Service
		expectedStatus int
		expectedBody   string
		called         bool
		apiKey         string
	}{
		{
			desc: "Test valid API key",
			mockTokenSvc: func() services.Service {
				mockTokenSvc := services.NewMockService(t)
				mockTokenSvc.On("GetToken", mock.Anything, "valid_api_key").Return(models.TokenData{
					APIKey:        "valid_api_key",
					RateLimit:     100,
					ExpiresAt:     validTimeStamp,
					AllowedRoutes: []string{"/api/v1/resource"},
				}, nil)
				mockTokenSvc.On("IncrementRateLimit", mock.Anything, mock.Anything).Return(int64(2), nil)
				return mockTokenSvc
			}(),
			expectedStatus: http.StatusOK,
			expectedBody:   `{"status":"success"}`,
			called:         true,
			apiKey:         "valid_api_key",
		},
		{
			desc: "Test missing API key",
			mockTokenSvc: func() services.Service {
				mockTokenSvc := services.NewMockService(t)
				return mockTokenSvc
			}(),
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized: API key is missing\n",
			called:         false,
		},
		{
			desc: "Test invalid API key",
			mockTokenSvc: func() services.Service {
				mockTokenSvc := services.NewMockService(t)
				mockTokenSvc.On("GetToken", mock.Anything, "invalid_api_key").Return(models.TokenData{}, redis.Nil)
				return mockTokenSvc
			}(),
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized: Invalid API key\n",
			called:         false,
			apiKey:         "invalid_api_key",
		},
		{
			desc: "Test invalid time format in token",
			mockTokenSvc: func() services.Service {
				mockTokenSvc := services.NewMockService(t)
				badTimeFormat := "bad_time_format"
				mockTokenSvc.On("GetToken", mock.Anything, "bad_time_format").Return(models.TokenData{
					APIKey:        "bad_time_format",
					RateLimit:     100,
					ExpiresAt:     badTimeFormat,
					AllowedRoutes: []string{"/api/v1/resource"},
				}, nil)
				return mockTokenSvc
			}(),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Internal Server Error: Invalid token expiry format\n",
			called:         false,
			apiKey:         "bad_time_format",
		},
		{
			desc: "Test expired token",
			mockTokenSvc: func() services.Service {
				mockTokenSvc := services.NewMockService(t)
				expiredTime := currentTime.Add(-24 * time.Hour).UTC().Format(time.RFC3339)
				mockTokenSvc.On("GetToken", mock.Anything, "expired_api_key").Return(models.TokenData{
					APIKey:        "expired_api_key",
					RateLimit:     100,
					ExpiresAt:     expiredTime,
					AllowedRoutes: []string{"/api/v1/resource"},
				}, nil)
				return mockTokenSvc
			}(),
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized: Invalid API key\n",
			called:         false,
			apiKey:         "expired_api_key",
		},
		{
			desc: "Test rate limit exceeded",
			mockTokenSvc: func() services.Service {
				mockTokenSvc := services.NewMockService(t)
				mockTokenSvc.On("GetToken", mock.Anything, "rate_limit_exceeded_api_key").Return(models.TokenData{
					APIKey:        "rate_limit_exceeded_api_key",
					RateLimit:     1,
					ExpiresAt:     validTimeStamp,
					AllowedRoutes: []string{"/api/v1/resource"},
				}, nil)
				mockTokenSvc.On("IncrementRateLimit", mock.Anything, mock.Anything).Return(int64(2), nil)
				return mockTokenSvc
			}(),
			expectedStatus: http.StatusTooManyRequests,
			expectedBody:   "Rate limit exceeded\n",
			called:         false,
			apiKey:         "rate_limit_exceeded_api_key",
		},
		{
			desc: "Test internal server error when getting token",
			mockTokenSvc: func() services.Service {
				mockTokenSvc := services.NewMockService(t)
				mockTokenSvc.On("GetToken", mock.Anything, "error_api_key").Return(models.TokenData{}, errors.New("some error"))
				return mockTokenSvc
			}(),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Internal Server Error\n",
			called:         false,
			apiKey:         "error_api_key",
		},
		{
			desc: "Test route not allowed for token",
			mockTokenSvc: func() services.Service {
				mockTokenSvc := services.NewMockService(t)
				mockTokenSvc.On("GetToken", mock.Anything, "not_allowed_api_key").Return(models.TokenData{
					APIKey:        "not_allowed_api_key",
					RateLimit:     100,
					ExpiresAt:     validTimeStamp,
					AllowedRoutes: []string{"/api/v1/other"},
				}, nil)
				return mockTokenSvc
			}(),
			expectedStatus: http.StatusForbidden,
			expectedBody:   "Route not allowed for this token\n",
			called:         false,
			apiKey:         "not_allowed_api_key",
		},
		{
			desc: "Test valid API key with no allowed routes",
			mockTokenSvc: func() services.Service {
				mockTokenSvc := services.NewMockService(t)
				mockTokenSvc.On("GetToken", mock.Anything, "no_routes_api_key").Return(models.TokenData{
					APIKey:        "no_routes_api_key",
					RateLimit:     100,
					ExpiresAt:     validTimeStamp,
					AllowedRoutes: []string{},
				}, nil)
				return mockTokenSvc
			}(),
			expectedStatus: http.StatusForbidden,
			expectedBody:   "Route not allowed for this token\n",
			called:         false,
			apiKey:         "no_routes_api_key",
		},
		{
			desc: "Test valid API key with multiple allowed routes",
			mockTokenSvc: func() services.Service {
				mockTokenSvc := services.NewMockService(t)
				mockTokenSvc.On("GetToken", mock.Anything, "multi_routes_api_key").Return(models.TokenData{
					APIKey:        "multi_routes_api_key",
					RateLimit:     100,
					ExpiresAt:     validTimeStamp,
					AllowedRoutes: []string{"/api/v1/resource", "/api/v1/another"},
				}, nil)
				mockTokenSvc.On("IncrementRateLimit", mock.Anything, mock.Anything).Return(int64(2), nil)
				return mockTokenSvc
			}(),
			expectedStatus: http.StatusOK,
			expectedBody:   `{"status":"success"}`,
			called:         true,
			apiKey:         "multi_routes_api_key",
		},
		{
			desc: "Test valid API key with rate limit reset",
			mockTokenSvc: func() services.Service {
				mockTokenSvc := services.NewMockService(t)
				mockTokenSvc.On("GetToken", mock.Anything, "reset_rate_limit_api_key").Return(models.TokenData{
					APIKey:        "reset_rate_limit_api_key",
					RateLimit:     100,
					ExpiresAt:     validTimeStamp,
					AllowedRoutes: []string{"/api/v1/resource"},
				}, nil)
				mockTokenSvc.On("IncrementRateLimit", mock.Anything, mock.Anything).Return(int64(0), nil)
				return mockTokenSvc
			}(),
			expectedStatus: http.StatusOK,
			expectedBody:   `{"status":"success"}`,
			called:         true,
			apiKey:         "reset_rate_limit_api_key",
		},
		{
			desc: "Test Internal Server Error when incrementing rate limit",
			mockTokenSvc: func() services.Service {
				mockTokenSvc := services.NewMockService(t)
				mockTokenSvc.On("GetToken", mock.Anything, "valid_api_key").Return(models.TokenData{
					APIKey:        "valid_api_key",
					RateLimit:     100,
					ExpiresAt:     validTimeStamp,
					AllowedRoutes: []string{"/api/v1/resource"},
				}, nil)
				mockTokenSvc.On("IncrementRateLimit", mock.Anything, mock.Anything).Return(int64(2), errors.New("some error"))
				return mockTokenSvc
			}(),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Internal Server Error\n",
			called:         false,
			apiKey:         "valid_api_key",
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			called := false
			mockService := new(services.MockService)

			defer mockService.AssertExpectations(t)
			req := httptest.NewRequest(http.MethodGet, "/api/v1/resource", nil)
			if tC.apiKey != "" {
				req.Header.Set("Authorization", "Bearer "+tC.apiKey)
			}
			rr := httptest.NewRecorder()
			finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"status":"success"}`))
				w.WriteHeader(http.StatusOK)
			})

			authM := ratelimit.RateLimitMiddleware{
				TokenService: tC.mockTokenSvc,
			}

			// Act
			authM.RateLimitHandler(finalHandler).ServeHTTP(rr, req)

			// Assert
			assert.Equal(t, tC.expectedStatus, rr.Code)
			assert.Equal(t, tC.expectedBody, rr.Body.String())
			if tC.called {
				assert.True(t, called, "Final handler should have been called")
			} else {
				assert.False(t, called, "Final handler should not have been called")
			}
		})
	}
}
