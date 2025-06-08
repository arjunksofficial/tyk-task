package auth_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/arjunksofficial/tyk-task/internal/middlewares/auth"
	"github.com/arjunksofficial/tyk-task/internal/token/models"
	"github.com/arjunksofficial/tyk-task/internal/token/services"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthMiddleware_AuthMiddleware(t *testing.T) {
	currentTime := time.Now()
	// expired1Hr := currentTime.Add(-time.Hour) // Token expired 1 hour ago
	// expired1HrTimeStamp := expired1Hr.UTC().Format(time.RFC3339)
	validTime := currentTime.Add(24 * time.Hour) // Token valid for 24 hours
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
			desc: "Test AuthMiddleware with valid API key",
			mockTokenSvc: func() services.Service {
				mockTokenSvc := services.NewMockService(t)
				mockTokenSvc.On("GetToken", mock.Anything, "valid_api_key").Return(models.TokenData{
					APIKey:        "valid_api_key",
					RateLimit:     100,
					ExpiresAt:     validTimeStamp,
					AllowedRoutes: []string{"/api/v1/resource"},
				}, nil)
				return mockTokenSvc
			}(),
			expectedStatus: http.StatusOK,
			expectedBody:   `{"status":"success"}`,
			called:         true,
			apiKey:         "valid_api_key",
		},
		{
			desc: "Test AuthMiddleware with missing API key",
			mockTokenSvc: func() services.Service {
				mockTokenSvc := services.NewMockService(t)
				return mockTokenSvc
			}(),
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized: API key is missing\n",
			called:         false,
		},
		{
			desc: "Test AuthMiddleware with invalid API key",
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
			desc: "Test AuthMiddleware with invalid API key",
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
			desc: "Test AuthMiddleware with invalid API key",
			mockTokenSvc: func() services.Service {
				mockTokenSvc := services.NewMockService(t)
				mockTokenSvc.On("GetToken", mock.Anything, "invalid_api_key").Return(models.TokenData{}, errors.New("some error"))
				return mockTokenSvc
			}(),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Internal Server Error\n",
			called:         false,
			apiKey:         "invalid_api_key",
		},
		{
			desc: "Test AuthMiddleware with expired API key",
			mockTokenSvc: func() services.Service {
				mockTokenSvc := services.NewMockService(t)
				mockTokenSvc.On("GetToken", mock.Anything, "expired_api_key").Return(models.TokenData{
					APIKey:        "expired_api_key",
					RateLimit:     100,
					ExpiresAt:     currentTime.Add(-time.Hour).UTC().Format(time.RFC3339),
					AllowedRoutes: []string{"/api/v1/resource"},
				}, nil)
				return mockTokenSvc
			}(),
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized: Invalid API key\n",
			called:         false,
			apiKey:         "expired_api_key",
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

			authM := auth.AuthMiddleware{
				TokenService: tC.mockTokenSvc,
			}

			// Act
			authM.AuthMiddleware(finalHandler).ServeHTTP(rr, req)

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
