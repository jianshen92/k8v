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

// LogOptions represents options for streaming pod logs
type LogOptions struct {
	TailLines    *int64
	HeadLines    *int64 // Limit to first N lines (not supported by K8s API, implemented by counting)
	SinceSeconds *int64
	Follow       bool
}

// StreamPodLogs streams logs from a specific pod container to the broadcast channel
func (c *Client) StreamPodLogs(
	ctx context.Context,
	namespace string,
	podName string,
	containerName string,
	opts LogOptions,
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
	logOptions := &corev1.PodLogOptions{
		Container:  containerName,
		Follow:     opts.Follow,
		Timestamps: true,
	}
	if opts.TailLines != nil {
		logOptions.TailLines = opts.TailLines
	}
	if opts.SinceSeconds != nil {
		logOptions.SinceSeconds = opts.SinceSeconds
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
	lineCount := int64(0)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case broadcast <- LogMessage{Type: "LOG_LINE", Line: scanner.Text() + "\n"}:
			// Sent successfully
			lineCount++
			// Stop if we've reached the head limit
			if opts.HeadLines != nil && lineCount >= *opts.HeadLines {
				broadcast <- LogMessage{Type: "LOG_END", Reason: fmt.Sprintf("Head limit reached (%d lines)", *opts.HeadLines)}
				return nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		broadcast <- LogMessage{Type: "LOG_ERROR", Error: err.Error()}
		return err
	}

	broadcast <- LogMessage{Type: "LOG_END", Reason: "EOF"}
	return nil
}
