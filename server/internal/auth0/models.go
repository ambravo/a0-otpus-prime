package auth0

import "time"

// ActionTrigger represents an Auth0 action trigger
type ActionTrigger struct {
	ID      string `json:"id"`
	Version string `json:"version"`
}

type Dependency struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	RegistryUrl string `json:"registry_url"`
}

type Secret struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// ActionCreate represents the payload for creating an Auth0 action
type ActionCreate struct {
	Name              string          `json:"name"`
	Code              string          `json:"code"`
	SupportedTriggers []ActionTrigger `json:"supported_triggers"`
	Runtime           string          `json:"runtime"`
	Secrets           []Secret        `json:"secrets"`
	Dependencies      []Dependency    `json:"dependencies"`
}

// ActionResponse represents the response from Auth0's action API
type ActionResponse struct {
	ID                string          `json:"id"`
	Name              string          `json:"name"`
	SupportedTriggers []ActionTrigger `json:"supported_triggers"`
	Status            string          `json:"status"`
	Created           string          `json:"created_at"`
	Updated           string          `json:"updated_at"`
}

type ActionBindings struct {
	Bindings []Binding `json:"bindings"`
}

type Binding struct {
	DisplayName string `json:"display_name"`
	ID          string `json:"id,omitempty"`
	Ref         Ref    `json:"ref"`
}

type Ref struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// ReadActionsResponse represents the response containing the total number of actions and a list of actions.
type ReadActionsResponse struct {
	Total   int              `json:"total"`
	Actions []ActionResponse `json:"actions"`
}

// TokenResponse represents the response from Auth0's token endpoint
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// DeviceCodeResponse represents the response from Auth0's device authorization endpoint
type DeviceCodeResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

type Providers struct {
	Providers []struct {
		Id            string `json:"id"`
		Tenant        string `json:"tenant"`
		Name          string `json:"name"`
		Channel       string `json:"channel"`
		Disabled      bool   `json:"disabled"`
		Configuration struct {
			DeliveryMethods interface{} `json:"delivery_methods"`
		} `json:"configuration"`
		Credentials interface{} `json:"credentials"`
		CreatedAt   time.Time   `json:"created_at"`
		UpdatedAt   time.Time   `json:"updated_at"`
	} `json:"providers"`
}

type UpdateProvider struct {
	Name          string `json:"name"`
	Disabled      bool   `json:"disabled"`
	Configuration struct {
		DeliveryMethods []string `json:"delivery_methods"`
	} `json:"configuration"`
}
