package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Arlan-Z/def-betting-api/internal/data"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

type PayoutClient interface {
	NotifyPayout(ctx context.Context, notification data.PayoutNotification) error
}

type RestyPayoutClient struct {
	client *resty.Client
	logger *zap.Logger
}

func NewRestyPayoutClient(baseURL string, timeout time.Duration, logger *zap.Logger) *RestyPayoutClient {
	client := resty.New().
		SetBaseURL(baseURL).
		SetTimeout(timeout).
		SetRetryCount(3).
		SetRetryWaitTime(100 * time.Millisecond).
		SetRetryMaxWaitTime(2 * time.Second).
		AddRetryCondition(func(r *resty.Response, err error) bool {
			return err != nil || r.StatusCode() >= http.StatusInternalServerError
		})

	client.OnBeforeRequest(func(c *resty.Client, r *resty.Request) error {
		// opentracing.GlobalTracer().Inject(...)
		return nil
	})

	client.OnError(func(req *resty.Request, err error) {
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
		logger.Error("Payout client request error",
			zap.String("method", req.Method),
			zap.String("url", req.URL),
			zap.Error(err),
			zap.Int("status_code", statusCode),
			zap.String("body", body))
	})

	return &RestyPayoutClient{
		client: client,
		logger: logger.Named("PayoutClient"), // Added logger name for clarity
	}
}

func (c *RestyPayoutClient) NotifyPayout(ctx context.Context, notification data.PayoutNotification) error {
	endpoint := "/payouts"

	req := c.client.R().SetContext(ctx)

	// if sessionID := ctx.Value("session_id"); sessionID != nil {
	//     req.SetHeader("Session-Id", sessionID.(string))
	// }

	bodyBytes, err := json.Marshal(notification)
	if err != nil {
		c.logger.Error("Failed to marshal payout notification", zap.Error(err))
		return fmt.Errorf("failed to marshal request body: %w", err)
	}
	req.SetBody(bytes.NewReader(bodyBytes))
	req.SetHeader("Content-Type", "application/json")

	resp, err := req.Post(endpoint)

	if err != nil {
		return fmt.Errorf("payout service request failed: %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("payout service returned error status %d", resp.StatusCode())
	}

	c.logger.Info("Successfully notified payout service",
		zap.String("userId", notification.UserID),
		zap.Float64("amount", notification.Amount))

	return nil
}
