package auth0

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

// Auth0Client represents an Auth0 API client
type Auth0Client struct {
	client *resty.Client
	logger *zap.Logger
}

var ErrAuthorizationPending = fmt.Errorf("authorization pending")

// NewAuth0Client creates a new Auth0 client
func NewAuth0Client() *Auth0Client {
	logger, _ := zap.NewProduction()

	client := resty.New().
		SetTimeout(10 * time.Second).
		SetRetryCount(3).
		SetRetryWaitTime(100 * time.Millisecond).
		SetRetryMaxWaitTime(2 * time.Second)

	return &Auth0Client{
		client: client,
		logger: logger,
	}
}

// InitiateDeviceFlow starts the device authorization flow
func (c *Auth0Client) InitiateDeviceFlow(domain string) (*DeviceCodeResponse, error) {
	resp, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]interface{}{
			"client_id": "2iZo3Uczt5LFHacKdM0zzgUO2eG2uDjT",
			"scope":     "create:actions update:actions",
			"audience":  fmt.Sprintf("https://%s/api/v2/", domain),
		}).
		Post(fmt.Sprintf("https://%s/oauth/device/code", domain))

	if err != nil {
		return nil, fmt.Errorf("device flow initiation failed: %w", err)
	}

	var response DeviceCodeResponse
	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		return nil, fmt.Errorf("failed to parse device code response: %w", err)
	}

	return &response, nil
}

// PollDeviceToken polls for the device token
func (c *Auth0Client) PollDeviceToken(domain, deviceCode string) (*TokenResponse, error) {
	resp, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]interface{}{
			"grant_type":  "urn:ietf:params:oauth:grant-type:device_code",
			"device_code": deviceCode,
			"client_id":   "2iZo3Uczt5LFHacKdM0zzgUO2eG2uDjT",
		}).
		Post(fmt.Sprintf("https://%s/oauth/token", domain))

	if err != nil {
		return nil, fmt.Errorf("token polling failed: %w", err)
	}

	if resp.StatusCode() == 403 {
		return nil, ErrAuthorizationPending
	}

	var response TokenResponse
	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	return &response, nil
}

// GetClientCredentialsToken gets a token using client credentials
func (c *Auth0Client) GetClientCredentialsToken(domain, clientID, clientSecret string) (*TokenResponse, error) {
	resp, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]interface{}{
			"client_id":     clientID,
			"client_secret": clientSecret,
			"audience":      fmt.Sprintf("https://%s/api/v2/", domain),
			"grant_type":    "client_credentials",
		}).
		Post(fmt.Sprintf("https://%s/oauth/token", domain))

	if err != nil {
		return nil, fmt.Errorf("client credentials flow failed: %w", err)
	}

	var response TokenResponse
	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	return &response, nil
}
