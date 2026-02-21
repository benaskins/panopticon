package aurelia

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// ServiceState mirrors the daemon's ServiceState for display.
type ServiceState struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	State        string `json:"state"`
	Health       string `json:"health"`
	PID          int    `json:"pid"`
	Port         int    `json:"port"`
	Uptime       string `json:"uptime"`
	RestartCount int    `json:"restart_count"`
	LastError    string `json:"last_error"`
}

// Client communicates with the Aurelia daemon over Unix socket.
type Client struct {
	socketPath string
	http       *http.Client
}

// NewClient creates a client targeting the default Aurelia socket.
func NewClient() *Client {
	socketPath := defaultSocketPath()
	return &Client{
		socketPath: socketPath,
		http: &http.Client{
			Timeout: 5 * time.Second,
			Transport: &http.Transport{
				DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
					return net.Dial("unix", socketPath)
				},
			},
		},
	}
}

// Available returns true if the Aurelia socket exists.
func (c *Client) Available() bool {
	_, err := os.Stat(c.socketPath)
	return err == nil
}

// Services fetches the current service states.
func (c *Client) Services() ([]ServiceState, error) {
	var states []ServiceState
	if err := c.get("/v1/services", &states); err != nil {
		return nil, err
	}
	return states, nil
}

// LogLines holds log output from a service.
type LogLines struct {
	Lines []string `json:"lines"`
}

// Logs fetches recent log lines for a service.
func (c *Client) Logs(name string, n int) ([]string, error) {
	var resp LogLines
	if err := c.get(fmt.Sprintf("/v1/services/%s/logs?n=%d", name, n), &resp); err != nil {
		return nil, err
	}
	return resp.Lines, nil
}

// ServiceAction sends a start/stop/restart action for a service.
func (c *Client) ServiceAction(name, action string) error {
	resp, err := c.http.Post(
		fmt.Sprintf("http://aurelia/v1/services/%s/%s", name, action),
		"application/json", nil,
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<16))
		return fmt.Errorf("API error %d: %s", resp.StatusCode, body)
	}
	return nil
}

func (c *Client) get(path string, v any) error {
	resp, err := c.http.Get("http://aurelia" + path)
	if err != nil {
		return fmt.Errorf("connecting to aurelia: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<16))
		return fmt.Errorf("API error %d: %s", resp.StatusCode, body)
	}

	return json.NewDecoder(resp.Body).Decode(v)
}

func defaultSocketPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".aurelia", "aurelia.sock")
}
