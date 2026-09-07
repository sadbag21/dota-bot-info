package dota

import (
	"context"
	"crypto/tls"
	"encoding/json"
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
		// Используем только IPv4
		DialContext: func(
			ctx context.Context,
			network string,
			address string,
		) (net.Conn, error) {
			return dialer.DialContext(ctx, "tcp4", address)
		},

		// Принудительно HTTP/1.1
		ForceAttemptHTTP2: false,

		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
			NextProtos: []string{"http/1.1"},
		},

		TLSHandshakeTimeout:   15 * time.Second,
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

func (c *Client) get(path string, result interface{}) error {
	url := c.baseURL + path

	req, err := http.NewRequest(
		http.MethodGet,
		url,
		nil,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to create request: %w",
			err,
		)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "dota-bot-info/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf(
			"request failed: %w",
			err,
		)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(
			"OpenDota returned status: %d",
			resp.StatusCode,
		)
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf(
			"failed to decode response: %w",
			err,
		)
	}

	return nil
}
