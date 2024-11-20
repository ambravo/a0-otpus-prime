package auth0

import (
	"embed"
	"encoding/json"
	"fmt"
	"github.com/ambravo/a0-OTPus-prime/server/internal/config"
	"github.com/ambravo/a0-OTPus-prime/server/internal/utils"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
	"log"
	"net/url"
	"time"
)

//go:embed actionTemplates
var templatesFS embed.FS

// EnablePhoneExtensibility creates or updates an Auth0 action
func (c *Auth0Client) EnablePhoneExtensibility(domain, accessToken string, chatID int64, cfg *config.Config) error {
	logger := c.logger
	var err error

	// Activate the Custom Phone Provider
	err = c.ActivateCustomPhoneProvider(domain, accessToken)

	// Custom Phone Provider, for Database Attributes
	err = c.UpdatePhoneActionTypeBased(domain, accessToken, chatID, cfg,
		"Custom Phone Provider", "custom-phone-provider", "v1")
	if err != nil {
		return fmt.Errorf("failed to Update action: %s", "Custom Phone Provider")
	}

	// Custom Phone Provider for MFA
	err = c.UpdatePhoneActionTypeBased(domain, accessToken, chatID, cfg,
		"Custom Phone Provider - MFA", "send-phone-message", "v2")
	if err != nil {
		logger.Error("AUTH0 is likely in a corrupt state!, please check actions and bindings")
		return fmt.Errorf("failed to Update action: %s", "Custom Phone Provider - MFA")
	}

	err = c.EnableMFA(domain, accessToken)
	if err != nil {
		return err
	}

	return nil
}

func (c *Auth0Client) UpdatePhoneActionTypeBased(domain string, accessToken string, chatID int64,
	cfg *config.Config, actionName string, actionType string, actionTypeVersion string) error {
	logger := c.logger

	postURL := fmt.Sprintf("%s/auth0/OTPs", cfg.BaseURL)
	tokenValue := fmt.Sprintf("%s:%d", domain, chatID)
	bearerToken := utils.GenerateAuth0DomainToken(tokenValue, cfg.HMACSecret)

	actionApiManagementURL := fmt.Sprintf("https://%s/api/v2/actions/actions", domain)

	var actionScriptSourceCode []byte
	var err error

	switch expr := actionType; expr {
	case "custom-phone-provider":
		actionScriptSourceCode, err = templatesFS.ReadFile("actionTemplates/onExecuteCustomPhoneProvider.js")
	case "send-phone-message":
		actionScriptSourceCode, err = templatesFS.ReadFile("actionTemplates/onExecuteSendPhoneMessage.js")
	default:
		return fmt.Errorf("unknown action type: %s", expr)
	}

	if err != nil {
		log.Fatalf("Failed to read Action SourceCode file: %v", err)
	}

	var secrets []Secret

	secrets = append(secrets, Secret{
		Name:  "BOT_GATEWAY_URL",
		Value: postURL,
	})
	secrets = append(secrets, Secret{
		Name:  "BOT_GATEWAY_TOKEN",
		Value: bearerToken,
	})
	secrets = append(secrets, Secret{
		Name:  "BOT_GATEWAY_CHAT_ID",
		Value: fmt.Sprintf("%d", chatID),
	})
	secrets = append(secrets, Secret{
		Name:  "AUTH0_DOMAIN",
		Value: domain,
	})

	action := ActionCreate{
		Name: actionName,
		Code: string(actionScriptSourceCode),
		SupportedTriggers: []ActionTrigger{
			{
				ID:      actionType,
				Version: actionTypeVersion,
			},
		},
		Runtime: "node18",
		Secrets: secrets,
		Dependencies: []Dependency{
			{
				Name:    "axios",
				Version: "1.7.7",
			},
		},
	}

	a0Client := c.client.R().SetAuthToken(accessToken).
		SetHeader("Content-Type", "application/json").
		SetBody(action)

	// First, try to get existing action
	existingAction, err := c.getAction(domain, accessToken, actionName)

	var responseBody []byte
	if err == nil && existingAction != nil {
		// Action exists, update it
		updateActionURL := actionApiManagementURL + "/" + existingAction.ID
		resp, err := a0Client.Patch(updateActionURL)

		if err != nil {
			return err
		}

		if resp.StatusCode() != 200 {
			return fmt.Errorf("failed to update action: %s", string(resp.Body()))
		}
		responseBody = resp.Body()
	} else if existingAction == nil {
		// Action does not exist, create it
		resp, err := a0Client.Post(actionApiManagementURL)

		if err != nil {
			logger.Error("Failed to create action",
				zap.Error(err),
				zap.String("domain", domain),
				zap.String("action", actionName))
			return fmt.Errorf("failed to create action: %w", err)
		}

		if resp.StatusCode() != 201 {
			return fmt.Errorf("failed to create action: %s", string(resp.Body()))
		}
		responseBody = resp.Body()
	}

	var actionResp ActionResponse
	if err := json.Unmarshal(responseBody, &actionResp); err != nil {
		return fmt.Errorf("failed to parse action response: %w", err)
	}
	// It is required to wait until the action changes from "Draft" to "Built" before it can be deployed
	var actionStatus = actionResp.Status
	for actionStatus != "built" {
		logger.Info("Waiting for action to be built",
			zap.String("domain", domain),
			zap.String("action", actionName),
			zap.String("status", actionStatus))
		time.Sleep(time.Millisecond * 1500)
		resp, err := c.getAction(domain, accessToken, actionName)
		if err != nil {
			return err
		}
		actionStatus = resp.Status
	}

	if err := c.deployAction(domain, accessToken, actionResp.ID); err != nil {
		logger.Error("Failed to deploy action", zap.String("domain", domain), zap.String("action", actionName))
		return err
	}

	return c.updateBindings(domain, accessToken, actionName, actionResp.ID, actionType)

}

func (c *Auth0Client) updateBindings(domain, accessToken, actionName, actionId string, actionType string) error {
	bindingsURL := fmt.Sprintf("https://%s/api/v2/actions/triggers/%s/bindings", domain, actionType)
	logger := c.logger

	// Fetch existing bindings
	resp, err := c.client.R().
		SetAuthToken(accessToken).
		Get(bindingsURL)
	if err != nil {
		logger.Error("Failed to fetch bindings", zap.Error(err))
		return fmt.Errorf("network error fetching bindings: %w", err)
	}

	var existingBindings ActionBindings
	if err := json.Unmarshal(resp.Body(), &existingBindings); err != nil {
		return fmt.Errorf("failed to parse bindings response: %w", err)
	}

	// Check if action is already bound
	var newBindings ActionBindings
	for _, binding := range existingBindings.Bindings {
		if binding.DisplayName == actionName {
			logger.Debug("Action already bound, removing it to refresh", zap.String("actionID", actionId))
		} else {
			newBindings.Bindings = append(newBindings.Bindings,
				Binding{
					DisplayName: binding.DisplayName,
					Ref: Ref{
						Type:  "binding_id",
						Value: binding.ID,
					},
				},
			)
		}
	}

	// Add new binding
	newBindings.Bindings = append(newBindings.Bindings,
		Binding{
			DisplayName: actionName,
			Ref: Ref{
				Type:  "action_id",
				Value: actionId,
			},
		},
	)

	// Update bindings
	updateResp, err := c.client.R().
		SetAuthToken(accessToken).
		SetHeader("Content-Type", "application/json").
		SetBody(newBindings).
		Patch(bindingsURL)
	if err != nil {
		logger.Error("Failed to update bindings", zap.Error(err))
		return fmt.Errorf("network error updating bindings: %w", err)
	}
	if updateResp.StatusCode() != 200 {
		logger.Error("Failed to add binding", zap.String("response", string(updateResp.Body())))
		return fmt.Errorf("HTTP error: %s", string(updateResp.Body()))
	}

	logger.Info("Binding added successfully", zap.String("actionID", actionId))
	return nil
}

func (c *Auth0Client) getAction(domain string, accessToken string, actionName string) (*ActionResponse, error) {
	apiManagementURL := fmt.Sprintf("https://%s/api/v2/actions/actions", domain)
	readActionURL := fmt.Sprintf("%s?actionName=%s", apiManagementURL, url.QueryEscape(actionName))

	resp, err := c.client.R().
		SetAuthToken(accessToken).
		Get(readActionURL)

	if err != nil {
		return nil, err
	}

	var readActions *ReadActionsResponse

	if err := json.Unmarshal(resp.Body(), &readActions); err != nil {
		return nil, err
	}

	if len(readActions.Actions) == 0 {
		return nil, fmt.Errorf("action not found")
	}

	return &readActions.Actions[0], nil
}

func (c *Auth0Client) deployAction(domain string, accessToken string, actionID string) error {
	logger := c.logger
	resp, err := c.client.R().
		SetAuthToken(accessToken).
		Post(fmt.Sprintf("https://%s/api/v2/actions/actions/%s/deploy", domain, actionID))

	if err != nil {
		return err
	}

	if resp.StatusCode() > 299 {
		return fmt.Errorf("failed to deploy action: %s", string(resp.Body()))
	}

	logger.Info("Action deployed", zap.String("actionID", actionID), zap.String("domain", domain))

	return nil
}
func (c *Auth0Client) ActivateCustomPhoneProvider(domain string, accessToken string) error {
	logger := c.logger
	var resp *resty.Response
	var err error
	resp, err = c.client.R().
		SetAuthToken(accessToken).
		Get(fmt.Sprintf("https://%s/api/v2/branding/phone/providers", domain))
	if err != nil {
		return err
	}

	var providers *Providers

	if err := json.Unmarshal(resp.Body(), &providers); err != nil {
		return err
	}

	if resp.StatusCode() > 299 {
		return fmt.Errorf("failed to read Phone provider: %s", string(resp.Body()))
	}

	// Patch providers and make sure the custom phone channel is enabled
	updateProvider := UpdateProvider{
		Name:     "custom",
		Disabled: false,
		Configuration: struct {
			DeliveryMethods []string `json:"delivery_methods"`
		}{
			DeliveryMethods: []string{"text"},
		},
	}

	resp, err = c.client.R().
		SetAuthToken(accessToken).
		SetBody(updateProvider).
		Patch(fmt.Sprintf("https://%s/api/v2/branding/phone/providers/%s", domain, providers.Providers[0].Id))

	if err != nil {
		return err
	}

	if resp.StatusCode() > 299 {
		return fmt.Errorf("failed to activate Phone provider: %s", string(resp.Body()))
	}

	logger.Info("Custom Provider Activated", zap.String("provider", providers.Providers[0].Id), zap.String("domain", domain))

	return nil
}

func (c *Auth0Client) EnableMFA(domain string, accessToken string) error {
	logger := c.logger
	var resp *resty.Response
	var err error

	// Step 1: Enable SMS factor
	resp, err = c.client.R().
		SetAuthToken(accessToken).
		SetBody(map[string]bool{"enabled": true}).
		Put(fmt.Sprintf("https://%s/api/v2/guardian/factors/sms", domain))
	if err != nil {
		return err
	}

	if resp.StatusCode() > 299 {
		return fmt.Errorf("failed to enable SMS factor: %s", string(resp.Body()))
	}

	// Step 2: Set the selected SMS provider to "phone-message-hook"
	resp, err = c.client.R().
		SetAuthToken(accessToken).
		SetBody(map[string]string{"provider": "phone-message-hook"}).
		Put(fmt.Sprintf("https://%s/api/v2/guardian/factors/phone/selected-provider", domain))
	if err != nil {
		return err
	}

	if resp.StatusCode() > 299 {
		return fmt.Errorf("failed to set SMS provider: %s", string(resp.Body()))
	}

	// Step 3: Set message types to ["sms", "voice"]
	resp, err = c.client.R().
		SetAuthToken(accessToken).
		SetBody(map[string][]string{"message_types": {"sms", "voice"}}).
		Put(fmt.Sprintf("https://%s/api/v2/guardian/factors/phone/message-types", domain))
	if err != nil {
		return err
	}

	if resp.StatusCode() > 299 {
		return fmt.Errorf("failed to set message types: %s", string(resp.Body()))
	}

	logger.Info("MFA enabled successfully", zap.String("domain", domain))

	return nil
}
