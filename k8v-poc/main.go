package main

// This program starts a small web server that:
//   1. Connects to a Kubernetes cluster using the local kubeconfig (usually ~/.kube/config).
//   2. Serves a static HTML UI at "/" from the local index.html file.
//   3. Exposes a WebSocket endpoint at "/ws" that, when connected,
//      streams live Kubernetes events (Pods, Deployments, ReplicaSets) to the browser.
// The frontend can then display these events in real time, giving a live view of
// what is happening in the cluster.

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/gorilla/websocket"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// WebSocket upgrader
// This is used to convert an incoming HTTP request into a persistent WebSocket
// connection, which we can then use to push JSON events to the browser.
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// ResourceEvent represents a generic Kubernetes resource event that is sent
// to the frontend over the WebSocket connection.
// It normalizes different resource types (Pods, Deployments, ReplicaSets)
// into a single, simple JSON shape that the UI can render.
type ResourceEvent struct {
	Type         string `json:"type"`         // "ADDED", "MODIFIED", "DELETED"
	ResourceType string `json:"resourceType"` // "Pod", "Deployment", "ReplicaSet"
	Name         string `json:"name"`
	Namespace    string `json:"namespace"`
	Status       string `json:"status"`
}

// main is the entry point. It:
//   - Parses flags (e.g. which port to listen on),
//   - Builds a Kubernetes client from the local kubeconfig,
//   - Registers HTTP handlers ("/" for the UI and "/ws" for WebSocket),
//   - Starts the HTTP server and blocks.
func main() {
	// Parse flags
	port := flag.String("port", "8080", "HTTP server port")
	flag.Parse()

	// Load kubeconfig from the user's home directory so that this binary
	// behaves like kubectl and talks to whatever cluster the user is
	// currently configured for.
	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatalf("Failed to load kubeconfig: %v", err)
	}

	// Create clientset (entry point to interact with the Kubernetes API).
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Failed to create clientset: %v", err)
	}

	fmt.Printf("Connected to Kubernetes cluster\n")

	// Set up HTTP routes:
	//   - "/" serves the static HTML UI.
	//   - "/ws" upgrades to WebSocket and streams cluster events.
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handleWebSocket(w, r, clientset)
	})

	// Start server
	addr := ":" + *port
	fmt.Printf("Starting server on http://localhost%s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// handleWebSocket upgrades the incoming HTTP request to a WebSocket connection,
// then starts three goroutines to watch Pods, Deployments, and ReplicaSets.
// Each watcher pushes events to the same WebSocket connection in a thread-safe way.
// When all watchers stop (e.g. due to error or connection close), the function returns
// and the WebSocket is closed.
func handleWebSocket(w http.ResponseWriter, r *http.Request, clientset *kubernetes.Clientset) {
	var mu sync.Mutex
	// Upgrade connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	defer conn.Close()

	fmt.Println("WebSocket client connected")

	// Use WaitGroup to keep connection open
	var wg sync.WaitGroup
	wg.Add(3)

	// Start watching resources in parallel.
	// Each watcher receives a shared WebSocket connection and mutex so that
	// concurrent writes to the socket do not race.
	go watchPods(conn, clientset, &wg, &mu)
	go watchDeployments(conn, clientset, &wg, &mu)
	go watchReplicaSets(conn, clientset, &wg, &mu)

	// Wait for all watchers
	wg.Wait()
	fmt.Println("WebSocket client disconnected")
}

// watchPods subscribes to all Pod events in all namespaces using the Kubernetes
// watch API and, for each event, sends a simplified ResourceEvent JSON object
// over the WebSocket connection.
// It exits if the watch fails to start or if writing to the WebSocket fails.
func watchPods(conn *websocket.Conn, clientset *kubernetes.Clientset, wg *sync.WaitGroup, mu *sync.Mutex) {
	defer wg.Done()

	// Background context is used here so the watch runs until:
	//   - The server process exits, or
	//   - The API server closes the watch.
	ctx := context.Background()
	watcher, err := clientset.CoreV1().Pods("").Watch(ctx, metav1.ListOptions{})
	if err != nil {
		log.Printf("Failed to watch pods: %v", err)
		return
	}
	defer watcher.Stop()

	fmt.Println("Started watching Pods")

	// Loop over events coming from the Kubernetes API and convert each one
	// into a ResourceEvent that the frontend can easily consume.
	for event := range watcher.ResultChan() {
		pod, ok := event.Object.(*v1.Pod)
		if !ok {
			continue
		}

		msg := ResourceEvent{
			Type:         string(event.Type),
			ResourceType: "Pod",
			Name:         pod.Name,
			Namespace:    pod.Namespace,
			Status:       string(pod.Status.Phase),
		}

		mu.Lock()
		if err := conn.WriteJSON(msg); err != nil {
			mu.Unlock()
			log.Printf("Failed to write pod event: %v", err)
			return
		}
		mu.Unlock()
	}
}

// watchDeployments is analogous to watchPods, but for Deployments.
// It converts Deployment watch events into ResourceEvent messages that
// include a human-readable "ready/total" replica count in the Status field.
func watchDeployments(conn *websocket.Conn, clientset *kubernetes.Clientset, wg *sync.WaitGroup, mu *sync.Mutex) {
	defer wg.Done()

	// Background context for the Deployment watch.
	ctx := context.Background()
	watcher, err := clientset.AppsV1().Deployments("").Watch(ctx, metav1.ListOptions{})
	if err != nil {
		log.Printf("Failed to watch deployments: %v", err)
		return
	}
	defer watcher.Stop()

	fmt.Println("Started watching Deployments")

	// Stream Deployment events from the API server to the browser.
	for event := range watcher.ResultChan() {
		deployment, ok := event.Object.(*appsv1.Deployment)
		if !ok {
			continue
		}

		status := fmt.Sprintf("%d/%d", deployment.Status.ReadyReplicas, deployment.Status.Replicas)

		msg := ResourceEvent{
			Type:         string(event.Type),
			ResourceType: "Deployment",
			Name:         deployment.Name,
			Namespace:    deployment.Namespace,
			Status:       status,
		}

		mu.Lock()
		if err := conn.WriteJSON(msg); err != nil {
			mu.Unlock()
			log.Printf("Failed to write deployment event: %v", err)
			return
		}
		mu.Unlock()
	}
}

// watchReplicaSets is similar to the other watcher functions but targets
// ReplicaSets. It is mainly useful for understanding the intermediate state
// between Deployments and Pods, since Deployments manage ReplicaSets, which
// in turn manage Pods.
func watchReplicaSets(conn *websocket.Conn, clientset *kubernetes.Clientset, wg *sync.WaitGroup, mu *sync.Mutex) {
	defer wg.Done()

	// Background context for the ReplicaSet watch.
	ctx := context.Background()
	watcher, err := clientset.AppsV1().ReplicaSets("").Watch(ctx, metav1.ListOptions{})
	if err != nil {
		log.Printf("Failed to watch replicasets: %v", err)
		return
	}
	defer watcher.Stop()

	fmt.Println("Started watching ReplicaSets")

	// Stream ReplicaSet events, skipping error events, to the WebSocket client.
	for event := range watcher.ResultChan() {
		rs, ok := event.Object.(*appsv1.ReplicaSet)
		if !ok {
			continue
		}

		// Skip if no event type (connection closed)
		if event.Type == watch.Error {
			continue
		}

		status := fmt.Sprintf("%d/%d", rs.Status.ReadyReplicas, rs.Status.Replicas)

		msg := ResourceEvent{
			Type:         string(event.Type),
			ResourceType: "ReplicaSet",
			Name:         rs.Name,
			Namespace:    rs.Namespace,
			Status:       status,
		}

		mu.Lock()
		if err := conn.WriteJSON(msg); err != nil {
			mu.Unlock()
			log.Printf("Failed to write replicaset event: %v", err)
			return
		}
		mu.Unlock()
	}
}
