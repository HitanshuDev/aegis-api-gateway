package gateway

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestNewAppliesTimeouts(t *testing.T) {
	timeouts := Timeouts{
		Read:       time.Second,
		ReadHeader: 2 * time.Second,
		Write:      3 * time.Second,
		Idle:       4 * time.Second,
	}
	server := New(":0", http.NotFoundHandler(), timeouts)

	if server.httpServer.ReadTimeout != timeouts.Read ||
		server.httpServer.ReadHeaderTimeout != timeouts.ReadHeader ||
		server.httpServer.WriteTimeout != timeouts.Write ||
		server.httpServer.IdleTimeout != timeouts.Idle {
		t.Fatalf("server timeouts = %+v, want %+v", server.httpServer, timeouts)
	}
}

func TestGracefulShutdownCompletesActiveRequests(t *testing.T) {
	requestStarted := make(chan struct{})
	releaseRequest := make(chan struct{})
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		close(requestStarted)
		<-releaseRequest
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "request completed")
	})

	server, listener, serveResult := startTestServer(t, handler)
	responseResult := make(chan error, 1)
	go func() {
		response, err := http.Get("http://" + listener.Addr().String())
		if err != nil {
			responseResult <- err
			return
		}
		defer response.Body.Close()

		body, err := io.ReadAll(response.Body)
		if err != nil {
			responseResult <- err
			return
		}
		if response.StatusCode != http.StatusOK {
			responseResult <- errors.New("unexpected response status")
			return
		}
		if string(body) != "request completed" {
			responseResult <- errors.New("unexpected response body")
			return
		}
		responseResult <- nil
	}()

	select {
	case <-requestStarted:
	case <-time.After(2 * time.Second):
		t.Fatal("request did not reach handler")
	}

	shutdownResult := make(chan error, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		shutdownResult <- server.Shutdown(ctx)
	}()

	select {
	case err := <-shutdownResult:
		t.Fatalf("Shutdown returned before active request completed: %v", err)
	case <-time.After(150 * time.Millisecond):
	}

	close(releaseRequest)
	if err := <-shutdownResult; err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}
	if err := <-responseResult; err != nil {
		t.Fatalf("active request failed: %v", err)
	}
	if err := <-serveResult; !errors.Is(err, http.ErrServerClosed) {
		t.Fatalf("Serve() error = %v, want %v", err, http.ErrServerClosed)
	}
}

func TestShutdownHonorsDeadlineAndCanForceClose(t *testing.T) {
	requestStarted := make(chan struct{})
	releaseRequest := make(chan struct{})
	handler := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		close(requestStarted)
		<-releaseRequest
	})

	server, listener, serveResult := startTestServer(t, handler)
	requestResult := make(chan error, 1)
	go func() {
		response, err := http.Get("http://" + listener.Addr().String())
		if err == nil {
			response.Body.Close()
		}
		requestResult <- err
	}()
	<-requestStarted

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	if err := server.Shutdown(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Shutdown() error = %v, want deadline exceeded", err)
	}
	if err := server.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if err := <-serveResult; !errors.Is(err, http.ErrServerClosed) {
		t.Fatalf("Serve() error = %v, want %v", err, http.ErrServerClosed)
	}

	close(releaseRequest)
	<-requestResult
}

func startTestServer(
	t *testing.T,
	handler http.Handler,
) (*Server, net.Listener, <-chan error) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := New(listener.Addr().String(), handler, Timeouts{})
	serveResult := make(chan error, 1)
	go func() {
		serveResult <- server.httpServer.Serve(listener)
	}()

	return server, listener, serveResult
}
