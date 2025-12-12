package k8s

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

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
	ExecMessageCreating  = "CREATING"  // Server -> Client: creating debug pod
	ExecMessageWaiting   = "WAITING"   // Server -> Client: waiting for pod ready
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

// NodeDebugPodOptions configures the debug pod for node shell access
type NodeDebugPodOptions struct {
	Image          string // Debug image (default: busybox:latest)
	Namespace      string // Namespace for debug pod (default: kube-system)
	TimeoutSeconds int    // Pod ready timeout (default: 120)
}

// DefaultNodeDebugPodOptions returns default options for node debug pods
func DefaultNodeDebugPodOptions() NodeDebugPodOptions {
	return NodeDebugPodOptions{
		Image:          "busybox:latest",
		Namespace:      "kube-system",
		TimeoutSeconds: 120,
	}
}

// CreateNodeDebugPod creates a privileged debug pod scheduled on the target node
// Returns the pod name and any error
func (c *Client) CreateNodeDebugPod(ctx context.Context, nodeName string, opts NodeDebugPodOptions) (string, error) {
	// Validate node exists
	_, err := c.Clientset.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("node not found: %w", err)
	}

	// Generate unique pod name
	podName := fmt.Sprintf("k8v-debug-%s-%d", nodeName, time.Now().Unix())

	// Create privileged pod spec
	privileged := true
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: opts.Namespace,
			Labels: map[string]string{
				"app":          "k8v-debug",
				"k8v.io/node":  nodeName,
				"k8v.io/debug": "true",
			},
		},
		Spec: corev1.PodSpec{
			NodeName:      nodeName, // Schedule on specific node
			HostPID:       true,
			HostNetwork:   true,
			HostIPC:       true,
			RestartPolicy: corev1.RestartPolicyNever,
			Containers: []corev1.Container{
				{
					Name:            "debug",
					Image:           opts.Image,
					ImagePullPolicy: corev1.PullIfNotPresent,
					Command:         []string{"sleep", "infinity"},
					SecurityContext: &corev1.SecurityContext{
						Privileged: &privileged,
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "host-root",
							MountPath: "/host",
						},
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: "host-root",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/",
						},
					},
				},
			},
		},
	}

	// Create the pod
	_, err = c.Clientset.CoreV1().Pods(opts.Namespace).Create(ctx, pod, metav1.CreateOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to create debug pod: %w", err)
	}

	c.logf("[NodeExec] Created debug pod %s/%s on node %s", opts.Namespace, podName, nodeName)
	return podName, nil
}

// DeleteNodeDebugPod deletes a debug pod
func (c *Client) DeleteNodeDebugPod(ctx context.Context, namespace, podName string) error {
	err := c.Clientset.CoreV1().Pods(namespace).Delete(ctx, podName, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete debug pod: %w", err)
	}
	c.logf("[NodeExec] Deleted debug pod %s/%s", namespace, podName)
	return nil
}

// WaitForPodReady waits for a pod to be running and ready
func (c *Client) WaitForPodReady(ctx context.Context, namespace, podName string, timeoutSeconds int) error {
	timeout := time.Duration(timeoutSeconds) * time.Second
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		pod, err := c.Clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get pod: %w", err)
		}

		// Check if pod is running
		if pod.Status.Phase == corev1.PodRunning {
			// Check if container is ready
			for _, containerStatus := range pod.Status.ContainerStatuses {
				if containerStatus.Name == "debug" && containerStatus.Ready {
					c.logf("[NodeExec] Debug pod %s/%s is ready", namespace, podName)
					return nil
				}
			}
		}

		// Check for failure
		if pod.Status.Phase == corev1.PodFailed {
			return fmt.Errorf("debug pod failed to start")
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(500 * time.Millisecond):
			// Continue polling
		}
	}

	return fmt.Errorf("timeout waiting for debug pod to be ready")
}

// ExecNodeDebugShell creates an interactive shell session in the debug pod
// It runs "chroot /host bash -l" to get full node access with a login shell
func (c *Client) ExecNodeDebugShell(
	ctx context.Context,
	namespace string,
	podName string,
	stdin io.Reader,
	stdout io.Writer,
	stderr io.Writer,
	sizeQueue remotecommand.TerminalSizeQueue,
) error {
	// Build exec request with chroot command
	// Use env to set TERM and HOME, then run bash as interactive login shell
	command := []string{
		"chroot", "/host",
		"/usr/bin/env",
		"TERM=xterm-256color",
		"HOME=/root",
		"/bin/bash", "--login",
	}

	req := c.Clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: "debug",
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

	c.logf("[NodeExec] Starting shell session in %s/%s", namespace, podName)

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
