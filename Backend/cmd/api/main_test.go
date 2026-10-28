package main

import (
	"context"
	"io"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestGracefulShutdownWaitsForInflightRequest(t *testing.T) {
	listener, e := net.Listen("tcp", "127.0.0.1:0")
	if e != nil {
		t.Fatal(e)
	}
	started := make(chan struct{})
	release := make(chan struct{})
	server := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-release
		_, _ = w.Write([]byte("concluído"))
	})}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- serve(ctx, server, listener) }()
	response := make(chan error, 1)
	go func() {
		client := http.Client{Timeout: 3 * time.Second}
		r, e := client.Get("http://" + listener.Addr().String())
		if e != nil {
			response <- e
			return
		}
		defer r.Body.Close()
		b, e := io.ReadAll(r.Body)
		if e == nil && string(b) != "concluído" {
			e = io.ErrUnexpectedEOF
		}
		response <- e
	}()
	select {
	case <-started:
	case <-time.After(3 * time.Second):
		t.Fatal("request did not start")
	}
	cancel()
	select {
	case e := <-done:
		t.Fatalf("shutdown returned before request finished: %v", e)
	case <-time.After(50 * time.Millisecond):
	}
	close(release)
	if e = <-response; e != nil {
		t.Fatal(e)
	}
	select {
	case e = <-done:
		if e != nil {
			t.Fatal(e)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("shutdown did not finish")
	}
	if conn, e := net.DialTimeout("tcp", listener.Addr().String(), 100*time.Millisecond); e == nil {
		conn.Close()
		t.Fatal("listener remained open")
	}
}
