package k8s

import (
	"bytes"
	"context"
	"fmt"
	"io"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

// ExecMessage represents a bidirectional exec communication message
type ExecMessage struct {
	Type string `json:"type"`           // INPUT, OUTPUT, RESIZE, CLOSE, ERROR, CONNECTED
	Data string `json:"data,omitempty"` // For INPUT/OUTPUT messages
	Cols uint16 `json:"cols,omitempty"` // For RESIZE messages
	Rows uint16 `json:"rows,omitempty"` // For RESIZE messages
}

// Exec message types
const (
	ExecMessageInput     = "INPUT"     // Client -> Server: keyboard input
	ExecMessageOutput    = "OUTPUT"    // Server -> Client: stdout/stderr
	ExecMessageResize    = "RESIZE"    // Client -> Server: terminal resize
	ExecMessageClose     = "CLOSE"     // Bidirectional: session ended
	ExecMessageError     = "ERROR"     // Server -> Client: error occurred
	ExecMessageConnected = "CONNECTED" // Server -> Client: shell ready
)

// TerminalSizeQueue implements remotecommand.TerminalSizeQueue
type TerminalSizeQueue struct {
	resizeChan chan remotecommand.TerminalSize
}

// NewTerminalSizeQueue creates a new terminal size queue
func NewTerminalSizeQueue() *TerminalSizeQueue {
	return &TerminalSizeQueue{
		resizeChan: make(chan remotecommand.TerminalSize, 1),
	}
}

// Next returns the next terminal size from the queue
func (q *TerminalSizeQueue) Next() *remotecommand.TerminalSize {
	size, ok := <-q.resizeChan
	if !ok {
		return nil
	}
	return &size
}

// Send sends a terminal size to the queue
func (q *TerminalSizeQueue) Send(cols, rows uint16) {
	// Non-blocking send - drop if queue is full
	select {
	case q.resizeChan <- remotecommand.TerminalSize{Width: cols, Height: rows}:
	default:
	}
}

// Close closes the resize channel
func (q *TerminalSizeQueue) Close() {
	close(q.resizeChan)
}

// DetectShell tries to detect an available shell in the container
// Returns /bin/bash if available, otherwise falls back to /bin/sh
func (c *Client) DetectShell(ctx context.Context, namespace, pod, container string) ([]string, error) {
	shells := [][]string{
		{"/bin/bash"},
		{"/bin/sh"},
	}

	for _, shell := range shells {
		// Test if shell exists by running a simple command
		req := c.Clientset.CoreV1().RESTClient().Post().
			Resource("pods").
			Name(pod).
			Namespace(namespace).
			SubResource("exec").
			VersionedParams(&corev1.PodExecOptions{
				Container: container,
				Command:   []string{"test", "-x", shell[0]},
				Stdin:     false,
				Stdout:    true,
				Stderr:    true,
				TTY:       false,
			}, scheme.ParameterCodec)

		exec, err := remotecommand.NewSPDYExecutor(c.config, "POST", req.URL())
		if err != nil {
			continue
		}

		var stdout, stderr bytes.Buffer
		err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
			Stdout: &stdout,
			Stderr: &stderr,
		})

		if err == nil {
			c.logf("[Exec] Detected shell: %s", shell[0])
			return shell, nil
		}
	}

	// Fallback to /bin/sh without testing
	c.logf("[Exec] Shell detection failed, falling back to /bin/sh")
	return []string{"/bin/sh"}, nil
}

// ExecPodShell creates an interactive shell session in a pod container
func (c *Client) ExecPodShell(
	ctx context.Context,
	namespace string,
	pod string,
	container string,
	command []string,
	stdin io.Reader,
	stdout io.Writer,
	stderr io.Writer,
	sizeQueue remotecommand.TerminalSizeQueue,
) error {
	// Validate pod exists
	podObj, err := c.Clientset.CoreV1().Pods(namespace).Get(ctx, pod, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("pod not found: %w", err)
	}

	// Validate container exists
	containerExists := false
	for _, c := range podObj.Spec.Containers {
		if c.Name == container {
			containerExists = true
			break
		}
	}
	if !containerExists {
		return fmt.Errorf("container not found: %s", container)
	}

	// Check pod is running
	if podObj.Status.Phase != corev1.PodRunning {
		return fmt.Errorf("pod is not running (status: %s)", podObj.Status.Phase)
	}

	// Build exec request
	req := c.Clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(pod).
		Namespace(namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: container,
			Command:   command,
			Stdin:     stdin != nil,
			Stdout:    stdout != nil,
			Stderr:    stderr != nil,
			TTY:       true,
		}, scheme.ParameterCodec)

	// Create SPDY executor
	exec, err := remotecommand.NewSPDYExecutor(c.config, "POST", req.URL())
	if err != nil {
		return fmt.Errorf("failed to create executor: %w", err)
	}

	// Stream with TTY support
	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:             stdin,
		Stdout:            stdout,
		Stderr:            stderr,
		Tty:               true,
		TerminalSizeQueue: sizeQueue,
	})

	if err != nil {
		return fmt.Errorf("exec stream error: %w", err)
	}

	return nil
}
