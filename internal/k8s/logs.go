package k8s

import (
	"bufio"
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// LogMessage represents a log streaming message
type LogMessage struct {
	Type   string `json:"type"`
	Line   string `json:"line,omitempty"`
	Reason string `json:"reason,omitempty"`
	Error  string `json:"error,omitempty"`
}

// StreamPodLogs streams logs from a specific pod container to the broadcast channel
func (c *Client) StreamPodLogs(
	ctx context.Context,
	namespace string,
	podName string,
	containerName string,
	broadcast chan<- LogMessage,
) error {
	// Validate pod exists first
	pod, err := c.Clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("pod not found: %w", err)
	}

	// Validate container exists
	containerExists := false
	for _, container := range pod.Spec.Containers {
		if container.Name == containerName {
			containerExists = true
			break
		}
	}
	if !containerExists {
		return fmt.Errorf("container not found: %s", containerName)
	}

	// Configure log options
	sinceSeconds := int64(600) // Last 10 minutes
	tailLines := int64(1000)   // Fallback line limit
	logOptions := &corev1.PodLogOptions{
		Container:    containerName,
		Follow:       true,
		Timestamps:   true,
		SinceSeconds: &sinceSeconds,
		TailLines:    &tailLines,
	}

	// Open log stream
	req := c.Clientset.CoreV1().Pods(namespace).GetLogs(podName, logOptions)
	stream, err := req.Stream(ctx)
	if err != nil {
		return fmt.Errorf("failed to open log stream: %w", err)
	}
	defer stream.Close()

	// Stream logs line by line
	scanner := bufio.NewScanner(stream)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case broadcast <- LogMessage{Type: "LOG_LINE", Line: scanner.Text() + "\n"}:
			// Sent successfully
		}
	}

	if err := scanner.Err(); err != nil {
		broadcast <- LogMessage{Type: "LOG_ERROR", Error: err.Error()}
		return err
	}

	broadcast <- LogMessage{Type: "LOG_END", Reason: "EOF"}
	return nil
}
