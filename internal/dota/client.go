package dota

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"
)

const baseURL = "https://api.opendota.com/api"

type Client struct {
	httpClient *http.Client
	baseURL    string
}

func NewClient() *Client {
	dialer := &net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	transport := &http.Transport{
		DialContext: func(
			ctx context.Context,
			network string,
			address string,
		) (net.Conn, error) {
			return dialer.DialContext(
				ctx,
				"tcp4",
				address,
			)
		},

		ForceAttemptHTTP2: false,

		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
			NextProtos: []string{
				"http/1.1",
			},
		},

		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
		IdleConnTimeout:       30 * time.Second,
	}

	return &Client{
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   40 * time.Second,
		},
		baseURL: baseURL,
	}
}

func (c *Client) get(
	path string,
	result interface{},
) error {
	url := c.baseURL + path

	req, err := http.NewRequest(
		http.MethodGet,
		url,
		nil,
	)
	if err != nil {
		return fmt.Errorf(
			"create request: %w",
			err,
		)
	}

	req.Header.Set(
		"Accept",
		"application/json",
	)

	req.Header.Set(
		"User-Agent",
		"dota-bot-info/1.0",
	)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return classifyResponseError(err, ErrUnavailable)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	switch {
	case resp.StatusCode == http.StatusOK:
		// всё хорошо

	case resp.StatusCode == http.StatusNotFound:
		return fmt.Errorf(
			"%w: status %d",
			ErrNotFound,
			resp.StatusCode,
		)

	case resp.StatusCode == http.StatusTooManyRequests:
		return fmt.Errorf(
			"%w: status %d",
			ErrRateLimited,
			resp.StatusCode,
		)

	case resp.StatusCode >= 500:
		return fmt.Errorf(
			"%w: status %d",
			ErrUnavailable,
			resp.StatusCode,
		)

	default:
		return fmt.Errorf(
			"%w: status %d",
			ErrBadResponse,
			resp.StatusCode,
		)
	}

	if err := json.NewDecoder(
		resp.Body,
	).Decode(result); err != nil {
		return fmt.Errorf("decode JSON: %w", classifyResponseError(err, ErrBadResponse))
	}

	return nil
}

// Timeouts can occur both before headers arrive and while reading the body.
func classifyResponseError(err, fallback error) error {
	var netErr net.Error
	if errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &netErr) && netErr.Timeout()) {
		return fmt.Errorf("%w: %w", ErrTimeout, err)
	}
	return fmt.Errorf("%w: %w", fallback, err)
}
