package main

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
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// ResourceEvent represents a K8s resource event
type ResourceEvent struct {
	Type         string `json:"type"`         // "ADDED", "MODIFIED", "DELETED"
	ResourceType string `json:"resourceType"` // "Pod", "Deployment", "ReplicaSet"
	Name         string `json:"name"`
	Namespace    string `json:"namespace"`
	Status       string `json:"status"`
}

func main() {
	// Parse flags
	port := flag.String("port", "8080", "HTTP server port")
	flag.Parse()

	// Load kubeconfig
	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatalf("Failed to load kubeconfig: %v", err)
	}

	// Create clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Failed to create clientset: %v", err)
	}

	fmt.Printf("Connected to Kubernetes cluster\n")

	// Set up HTTP routes
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

	// Start watching resources
	go watchPods(conn, clientset, &wg, &mu)
	go watchDeployments(conn, clientset, &wg, &mu)
	go watchReplicaSets(conn, clientset, &wg, &mu)

	// Wait for all watchers
	wg.Wait()
	fmt.Println("WebSocket client disconnected")
}

func watchPods(conn *websocket.Conn, clientset *kubernetes.Clientset, wg *sync.WaitGroup, mu *sync.Mutex) {
	defer wg.Done()

	ctx := context.Background()
	watcher, err := clientset.CoreV1().Pods("").Watch(ctx, metav1.ListOptions{})
	if err != nil {
		log.Printf("Failed to watch pods: %v", err)
		return
	}
	defer watcher.Stop()

	fmt.Println("Started watching Pods")

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

func watchDeployments(conn *websocket.Conn, clientset *kubernetes.Clientset, wg *sync.WaitGroup, mu *sync.Mutex) {
	defer wg.Done()

	ctx := context.Background()
	watcher, err := clientset.AppsV1().Deployments("").Watch(ctx, metav1.ListOptions{})
	if err != nil {
		log.Printf("Failed to watch deployments: %v", err)
		return
	}
	defer watcher.Stop()

	fmt.Println("Started watching Deployments")

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

func watchReplicaSets(conn *websocket.Conn, clientset *kubernetes.Clientset, wg *sync.WaitGroup, mu *sync.Mutex) {
	defer wg.Done()

	ctx := context.Background()
	watcher, err := clientset.AppsV1().ReplicaSets("").Watch(ctx, metav1.ListOptions{})
	if err != nil {
		log.Printf("Failed to watch replicasets: %v", err)
		return
	}
	defer watcher.Stop()

	fmt.Println("Started watching ReplicaSets")

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
