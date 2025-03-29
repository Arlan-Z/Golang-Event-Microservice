package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/Arlan-Z/def-betting-api/internal/domain"
)

const defaultTimeout = 10 * time.Second

// Client is used to interact with the external betting API.
type Client struct {
	httpClient *http.Client
	baseURL    *url.URL
}

func NewClient(baseURL string) (*Client, error) {
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL '%s': %w", baseURL, err)
	}

	return &Client{
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		baseURL: parsedURL,
	}, nil
}

func (c *Client) buildURL(pathSegments ...string) string {
	finalURL := *c.baseURL
	finalURL.Path = c.baseURL.Path + "/" + url.PathEscape(pathSegments[0])
	if len(pathSegments) > 1 {
		additionalPath := ""
		for _, segment := range pathSegments[1:] {
			additionalPath += "/" + url.PathEscape(segment)
		}
		finalURL.Path += additionalPath
	}
	return finalURL.String()
}

func (c *Client) doRequest(ctx context.Context, method, urlStr string, reqBody, respBody interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if reqBody != nil {
		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, urlStr, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request to %s: %w", urlStr, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return resp, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	if respBody != nil {
		if _, ok := respBody.(*domain.SubscriptionResponse); ok && resp.Header.Get("Content-Type") != "application/json" {
			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				return resp, fmt.Errorf("failed to read plain text response body: %w", err)
			}
			*respBody.(*domain.SubscriptionResponse) = domain.SubscriptionResponse(bodyBytes)
			return resp, nil
		}

		if err := json.NewDecoder(resp.Body).Decode(respBody); err != nil {
			return resp, fmt.Errorf("failed to decode response body: %w", err)
		}
	}

	return resp, nil
}

func (c *Client) GetAllEvents(ctx context.Context) ([]domain.Event, error) {
	urlStr := c.buildURL("all")
	var events []domain.Event
	_, err := c.doRequest(ctx, http.MethodGet, urlStr, nil, &events)
	if err != nil {
		return nil, fmt.Errorf("GetAllEvents failed: %w", err)
	}
	return events, nil
}

func (c *Client) GetEvent(ctx context.Context, eventID string) (*domain.Event, error) {
	if eventID == "" {
		return nil, fmt.Errorf("GetEvent failed: eventID cannot be empty")
	}
	urlStr := c.buildURL(eventID)
	var event domain.Event
	_, err := c.doRequest(ctx, http.MethodGet, urlStr, nil, &event)
	if err != nil {
		return nil, fmt.Errorf("GetEvent failed for ID %s: %w", eventID, err)
	}
	return &event, nil
}

func (c *Client) GetEventDetails(ctx context.Context, eventID string) (*domain.EventDetails, error) {
	if eventID == "" {
		return nil, fmt.Errorf("GetEventDetails failed: eventID cannot be empty")
	}
	urlStr := c.buildURL(eventID, "details")
	var details domain.EventDetails
	_, err := c.doRequest(ctx, http.MethodGet, urlStr, nil, &details)
	if err != nil {
		return nil, fmt.Errorf("GetEventDetails failed for ID %s: %w", eventID, err)
	}
	return &details, nil
}

func (c *Client) SubscribeToEvent(ctx context.Context, eventID string, callbackURL string) (domain.SubscriptionResponse, error) {
	if eventID == "" {
		return "", fmt.Errorf("SubscribeToEvent failed: eventID cannot be empty")
	}
	if callbackURL == "" {
		return "", fmt.Errorf("SubscribeToEvent failed: callbackURL cannot be empty")
	}

	urlStr := c.buildURL(eventID, "subscribe")
	reqBody := domain.SubscriptionRequest{CallbackURL: callbackURL}
	var respBody domain.SubscriptionResponse

	_, err := c.doRequest(ctx, http.MethodPost, urlStr, reqBody, &respBody)
	if err != nil {
		return "", fmt.Errorf("SubscribeToEvent failed for ID %s: %w", eventID, err)
	}

	return respBody, nil
}
