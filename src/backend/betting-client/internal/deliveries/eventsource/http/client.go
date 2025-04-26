package http

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Arlan-Z/def-betting-api/internal/data"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

type EventSourceClient interface {
	FetchActiveEvents(ctx context.Context) ([]data.ExternalEventDTO, error)
}

type RestyEventSourceClient struct {
	client *resty.Client
	logger *zap.Logger
}

func NewRestyEventSourceClient(baseURL string, timeout time.Duration, logger *zap.Logger) *RestyEventSourceClient {
	client := resty.New().
		SetBaseURL(baseURL).
		SetTimeout(timeout).
		OnError(func(req *resty.Request, err error) {
			var resp *resty.Response
			if r, ok := err.(*resty.ResponseError); ok {
				resp = r.Response
			}
			statusCode := 0
			body := ""
			if resp != nil {
				statusCode = resp.StatusCode()
				body = string(resp.Body())
			}
			logger.Error("Event source client request error",
				zap.String("method", req.Method),
				zap.String("url", req.URL),
				zap.Error(err),
				zap.Int("status_code", statusCode),
				zap.String("body", body))
		})

	return &RestyEventSourceClient{
		client: client,
		logger: logger.Named("EventSourceClient"),
	}
}

func (c *RestyEventSourceClient) FetchActiveEvents(ctx context.Context) ([]data.ExternalEventDTO, error) {
	endpoint := "/api/v1/events/active"

	req := c.client.R().SetContext(ctx)

	resp, err := req.Get(endpoint)

	if err != nil {
		c.logger.Error("Error requesting events from source API", zap.Error(err))
		return nil, fmt.Errorf("failed to execute request to event source: %w", err)
	}

	if resp.IsError() {
		c.logger.Error("Event source API returned an error",
			zap.Int("status_code", resp.StatusCode()),
			zap.String("body", string(resp.Body())),
		)
		return nil, fmt.Errorf("event source returned status %d", resp.StatusCode())
	}

	var events []data.ExternalEventDTO
	err = json.Unmarshal(resp.Body(), &events)
	if err != nil {
		c.logger.Error("Error decoding response from event source API", zap.Error(err))
		return nil, fmt.Errorf("failed to decode response from event source: %w", err)
	}

	c.logger.Debug("Successfully fetched events from external API", zap.Int("count", len(events)))
	return events, nil
}
