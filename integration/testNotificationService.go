package integration

import (
	"bufio"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type LogConsumer struct{}

func (lc *LogConsumer) Accept(l testcontainers.Log) {
	fmt.Printf("[NotificationTestContainer] %s\n", string(l.Content))
}

type NotificationContainer struct {
	container testcontainers.Container
	ctx       context.Context
	port      string
}

func NewNotificationContainer(ctx context.Context) (*NotificationContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "greenbone/exercise-admin-notification",
		ExposedPorts: []string{"8080/tcp"},
		WaitingFor:   wait.ForLog("Starting Admin Notification service"),
		LogConsumerCfg: &testcontainers.LogConsumerConfig{
			Opts:      []testcontainers.LogProductionOption{testcontainers.WithLogProductionTimeout(10 * time.Second)},
			Consumers: []testcontainers.LogConsumer{&LogConsumer{}},
		},
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	port, err := container.MappedPort(ctx, "8080/tcp")
	if err != nil {
		return nil, fmt.Errorf("failed to get container port: %w", err)
	}

	return &NotificationContainer{
		container: container,
		ctx:       ctx,
		port:      port.Port(),
	}, nil
}

func (nc *NotificationContainer) GetNotificationURL() string {
	return fmt.Sprintf("http://localhost:%s/api/notify", nc.port)
}

func (nc *NotificationContainer) Terminate() error {
	return nc.container.Terminate(nc.ctx)
}

func (nc *NotificationContainer) SearchLogs(pattern string) ([]string, error) {
	logs, err := nc.container.Logs(nc.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get container logs: %w", err)
	}
	defer logs.Close()

	var matches []string
	scanner := bufio.NewScanner(logs)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, pattern) {
			matches = append(matches, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read logs: %w", err)
	}

	return matches, nil
}

func (nc *NotificationContainer) WaitForLog(pattern string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		matches, err := nc.SearchLogs(pattern)
		if err != nil {
			return err
		}
		if len(matches) > 0 {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for log pattern: %s", pattern)
}
