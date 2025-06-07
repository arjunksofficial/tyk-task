package models

import (
	"encoding/json"
	"time"
)

type TokenData struct {
	APIKey        string   `json:"api_key"`
	RateLimit     int      `json:"rate_limit"`
	ExpiresAt     string   `json:"expires_at"`
	AllowedRoutes []string `json:"allowed_routes"`
}

type ContextKey string

const (
	TokenContextKey ContextKey = "token"
)

func NewToken(apiKey string) *TokenData {
	return &TokenData{
		APIKey:        apiKey,
		RateLimit:     100, // Default rate limit
		ExpiresAt:     "",  // To be set later
		AllowedRoutes: []string{"/api/v1/users/*", "/api/v1/products/*"},
	}
}

func (t *TokenData) SetExpiry(duration int64) {
	// Set the expiry time in RFC3339 format
	t.ExpiresAt = time.Now().Add(time.Duration(duration) * time.Second).UTC().Format(time.RFC3339)
}

func (t *TokenData) SetRateLimit(limit int) {
	t.RateLimit = limit
}

// SetAllowedRoutes sets the allowed routes for the token
func (t *TokenData) SetAllowedRoutes(routes []string) {
	t.AllowedRoutes = routes
}

// IsExpired checks if the token is expired based on the ExpiresAt field
func (t *TokenData) IsExpired() bool {
	if t.ExpiresAt == "" {
		return false // If no expiry is set, consider it not expired
	}
	expiryTime, err := time.Parse(time.RFC3339, t.ExpiresAt)
	if err != nil {
		return false // If parsing fails, consider it not expired
	}
	return time.Now().After(expiryTime)
}

// IsValidRoute checks if the given route is allowed by the token
func (t *TokenData) IsValidRoute(route string) bool {
	for _, allowedRoute := range t.AllowedRoutes {
		if allowedRoute == route || allowedRoute == "*" || (allowedRoute[len(allowedRoute)-1] == '*' && route[:len(allowedRoute)-1] == allowedRoute[:len(allowedRoute)-1]) {
			return true
		}
	}
	return false
}

// IsValidRateLimit checks if the request count is within the token's rate limit
func (t *TokenData) IsValidRateLimit(requestCount int) bool {
	return requestCount <= t.RateLimit
}

// ToJSON converts the TokenData to a JSON string
func (t *TokenData) ToJSON() (string, error) {
	data, err := json.Marshal(t)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// TokenFromJSON converts a JSON string to a TokenData object
// This function is useful for deserializing token data from a JSON string
func TokenFromJSON(data string) (*TokenData, error) {
	var token TokenData
	err := json.Unmarshal([]byte(data), &token)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// String returns a string representation of the token data
func (t *TokenData) String() string {
	data, err := t.ToJSON()
	if err != nil {
		return "Invalid Token Data"
	}
	return data
}

// IsValid checks if the token is valid based on its expiry and rate limit
func (t *TokenData) IsValid() bool {
	return !t.IsExpired() && t.IsValidRateLimit(0) // Replace 0 with actual request count
}

// IsValid checks if the token is valid based on its expiry and rate limit
func (t *TokenData) IsValidWithRequestCount(requestCount int) bool {
	return !t.IsExpired() && t.IsValidRateLimit(requestCount)
}

// IsValid checks if the token is valid based on its expiry and rate limit
func (t *TokenData) IsValidWithRoute(route string) bool {
	return !t.IsExpired() && t.IsValidRoute(route)
}

// IsValid checks if the token is valid based on its expiry, rate limit, and route
func (t *TokenData) IsValidWithRequestCountAndRoute(requestCount int, route string) bool {
	return !t.IsExpired() && t.IsValidRateLimit(requestCount) && t.IsValidRoute(route)
}
