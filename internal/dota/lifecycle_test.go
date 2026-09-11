package dota

import (
	"context"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"
)

type canceledBody struct {
	ctx     context.Context
	started chan struct{}
}

func (b canceledBody) Read([]byte) (int, error) {
	close(b.started)
	<-b.ctx.Done()
	return 0, b.ctx.Err()
}
func (b canceledBody) Close() error { return nil }

func TestOpenDotaShutdownCancelsRequestAndBody(t *testing.T) {
	for _, phase := range []string{"request", "body"} {
		t.Run(phase, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			started := make(chan struct{})
			client := NewClientWithContext(ctx)
			defer client.Close()
			client.httpClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if phase == "request" {
					close(started)
					<-req.Context().Done()
					return nil, req.Context().Err()
				}
				return &http.Response{StatusCode: 200, Body: io.ReadCloser(canceledBody{req.Context(), started}), Header: make(http.Header)}, nil
			})
			done := make(chan error, 1)
			go func() { _, err := client.GetPlayer(123); done <- err }()
			select {
			case <-started:
			case <-time.After(2 * time.Second):
				t.Fatal("request did not start")
			}
			cancel()
			select {
			case err := <-done:
				if !errors.Is(err, context.Canceled) {
					t.Fatalf("cancellation lost: %v", err)
				}
			case <-time.After(time.Second):
				t.Fatal("request not canceled")
			}
		})
	}
}
