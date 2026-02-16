package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"gotalk/internal/models"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	addr, err := pickFreeAddr()
	if err != nil {
		exitf("failed to pick free local port: %v", err)
	}

	srvCmd := exec.CommandContext(ctx, "go", "run", "./cmd/server", "-addr", addr)
	var logBuf bytes.Buffer
	srvCmd.Stdout = io.MultiWriter(io.Discard, &logBuf)
	srvCmd.Stderr = io.MultiWriter(io.Discard, &logBuf)

	if err := srvCmd.Start(); err != nil {
		exitf("failed to start server: %v", err)
	}
	defer stopServer(srvCmd)

	serverDone := make(chan error, 1)
	go func() {
		serverDone <- srvCmd.Wait()
	}()

	if err := waitForHTTP(ctx, "http://"+addr+"/", serverDone); err != nil {
		if logs := strings.TrimSpace(logBuf.String()); logs != "" {
			exitf("server did not become ready: %v\nserver output:\n%s", err, logs)
		}
		exitf("server did not become ready: %v", err)
	}

	receiver, err := dial(addr, "smoke", "receiver")
	if err != nil {
		exitf("receiver dial failed: %v", err)
	}
	defer receiver.Close()

	sender, err := dial(addr, "smoke", "sender")
	if err != nil {
		exitf("sender dial failed: %v", err)
	}
	defer sender.Close()

	if err := sender.WriteMessage(websocket.TextMessage, []byte("smoke-test")); err != nil {
		exitf("sender write failed: %v", err)
	}

	if err := expectMessage(receiver, "sender", "smoke-test", 5*time.Second); err != nil {
		exitf("message assertion failed: %v", err)
	}

	fmt.Println("PASS: websocket E2E smoke test succeeded")
}

func dial(addr, room, user string) (*websocket.Conn, error) {
	u := url.URL{Scheme: "ws", Host: addr, Path: "/ws"}
	q := u.Query()
	q.Set("room", room)
	q.Set("user", user)
	u.RawQuery = q.Encode()
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	return conn, err
}

func expectMessage(conn *websocket.Conn, wantUser, wantContent string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		_ = conn.SetReadDeadline(time.Now().Add(1200 * time.Millisecond))
		_, payload, err := conn.ReadMessage()
		if err != nil {
			if isTimeout(err) {
				continue
			}
			return err
		}

		var msg models.Message
		if err := json.Unmarshal(payload, &msg); err != nil {
			return fmt.Errorf("invalid JSON payload: %w", err)
		}

		if msg.Type == models.TypeMessage && msg.User == wantUser && msg.Content == wantContent {
			return nil
		}
	}
	return errors.New("timed out waiting for expected message")
}

func waitForHTTP(ctx context.Context, endpoint string, serverDone <-chan error) error {
	client := &http.Client{Timeout: 500 * time.Millisecond}
	ticker := time.NewTicker(120 * time.Millisecond)
	defer ticker.Stop()

	for {
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		resp, err := client.Do(req)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}

		select {
		case err := <-serverDone:
			if err == nil {
				return errors.New("server exited before readiness check completed")
			}
			return fmt.Errorf("server exited early: %w", err)
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func pickFreeAddr() (string, error) {
	l, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return "", err
	}
	defer l.Close()
	return l.Addr().String(), nil
}

func stopServer(cmd *exec.Cmd) {
	if cmd.Process == nil {
		return
	}
	_ = cmd.Process.Signal(os.Interrupt)
	time.Sleep(300 * time.Millisecond)
	_ = cmd.Process.Kill()
}

func isTimeout(err error) bool {
	if err == nil {
		return false
	}
	if ne, ok := err.(interface{ Timeout() bool }); ok && ne.Timeout() {
		return true
	}
	return strings.Contains(strings.ToLower(err.Error()), "timeout")
}

func exitf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "FAIL: "+format+"\n", args...)
	os.Exit(1)
}
