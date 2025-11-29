package k8s

import (
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"sigs.k8s.io/yaml"

	"github.com/user/k8v/internal/types"
)

// TransformPod converts a Kubernetes Pod to our Resource model
func TransformPod(pod *v1.Pod, cache *ResourceCache) *types.Resource {
	podID := types.BuildID("Pod", pod.Namespace, pod.Name)

	resource := &types.Resource{
		ID:        podID,
		Type:      "Pod",
		Name:      pod.Name,
		Namespace: pod.Namespace,

		Status: types.ResourceStatus{
			Phase:   string(pod.Status.Phase),
			Ready:   getPodReadyStatus(pod),
			Message: getPodMessage(pod),
		},

		Health: computePodHealth(pod),

		Relationships: types.Relationships{
			OwnedBy:     ExtractOwners(pod),
			DependsOn:   append(ExtractConfigMapDeps(pod), ExtractSecretDeps(pod)...),
			ExposedBy:   FindReverseRelationships(podID, types.RelExposes, cache),
			ScheduledOn: ExtractPodNodeScheduling(pod),
		},

		Labels:      pod.Labels,
		Annotations: pod.Annotations,
		CreatedAt:   pod.CreationTimestamp.Time,
		Spec:        pod.Spec,
		YAML:        marshalToYAML(pod),
	}

	return resource
}

// TransformDeployment converts a Kubernetes Deployment to our Resource model
func TransformDeployment(deployment *appsv1.Deployment, cache *ResourceCache) *types.Resource {
	deploymentID := types.BuildID("Deployment", deployment.Namespace, deployment.Name)

	resource := &types.Resource{
		ID:        deploymentID,
		Type:      "Deployment",
		Name:      deployment.Name,
		Namespace: deployment.Namespace,

		Status: types.ResourceStatus{
			Phase:   getDeploymentPhase(deployment),
			Ready:   fmt.Sprintf("%d/%d", deployment.Status.ReadyReplicas, deployment.Status.Replicas),
			Message: getDeploymentMessage(deployment),
		},

		Health: computeDeploymentHealth(deployment),

		Relationships: types.Relationships{
			OwnedBy: ExtractOwners(deployment),
			Owns:    FindReverseRelationships(deploymentID, types.RelOwnedBy, cache),
		},

		Labels:      deployment.Labels,
		Annotations: deployment.Annotations,
		CreatedAt:   deployment.CreationTimestamp.Time,
		Spec:        deployment.Spec,
		YAML:        marshalToYAML(deployment),
	}

	return resource
}

// TransformReplicaSet converts a Kubernetes ReplicaSet to our Resource model
func TransformReplicaSet(rs *appsv1.ReplicaSet, cache *ResourceCache) *types.Resource {
	rsID := types.BuildID("ReplicaSet", rs.Namespace, rs.Name)

	resource := &types.Resource{
		ID:        rsID,
		Type:      "ReplicaSet",
		Name:      rs.Name,
		Namespace: rs.Namespace,

		Status: types.ResourceStatus{
			Phase:   "Active",
			Ready:   fmt.Sprintf("%d/%d", rs.Status.ReadyReplicas, rs.Status.Replicas),
			Message: "",
		},

		Health: computeReplicaSetHealth(rs),

		Relationships: types.Relationships{
			OwnedBy: ExtractOwners(rs),
			Owns:    FindReverseRelationships(rsID, types.RelOwnedBy, cache),
		},

		Labels:      rs.Labels,
		Annotations: rs.Annotations,
		CreatedAt:   rs.CreationTimestamp.Time,
		Spec:        rs.Spec,
		YAML:        marshalToYAML(rs),
	}

	return resource
}

// TransformService converts a Kubernetes Service to our Resource model
func TransformService(service *v1.Service, cache *ResourceCache) *types.Resource {
	serviceID := types.BuildID("Service", service.Namespace, service.Name)

	resource := &types.Resource{
		ID:        serviceID,
		Type:      "Service",
		Name:      service.Name,
		Namespace: service.Namespace,

		Status: types.ResourceStatus{
			Phase:   "Active",
			Ready:   "",
			Message: "",
		},

		Health: types.HealthHealthy,

		Relationships: types.Relationships{
			OwnedBy:  ExtractOwners(service),
			Exposes:  FindExposedPods(service, cache),
			RoutedBy: FindReverseRelationships(serviceID, types.RelRoutesTo, cache),
		},

		Labels:      service.Labels,
		Annotations: service.Annotations,
		CreatedAt:   service.CreationTimestamp.Time,
		Spec:        service.Spec,
		YAML:        marshalToYAML(service),
	}

	return resource
}

// TransformIngress converts a Kubernetes Ingress to our Resource model
func TransformIngress(ingress *netv1.Ingress, cache *ResourceCache) *types.Resource {
	resource := &types.Resource{
		ID:        types.BuildID("Ingress", ingress.Namespace, ingress.Name),
		Type:      "Ingress",
		Name:      ingress.Name,
		Namespace: ingress.Namespace,

		Status: types.ResourceStatus{
			Phase:   "Active",
			Ready:   "",
			Message: "",
		},

		Health: types.HealthHealthy,

		Relationships: types.Relationships{
			OwnedBy:  ExtractOwners(ingress),
			RoutesTo: FindRoutedServices(ingress),
		},

		Labels:      ingress.Labels,
		Annotations: ingress.Annotations,
		CreatedAt:   ingress.CreationTimestamp.Time,
		Spec:        ingress.Spec,
		YAML:        marshalToYAML(ingress),
	}

	return resource
}

// TransformConfigMap converts a Kubernetes ConfigMap to our Resource model
func TransformConfigMap(cm *v1.ConfigMap, cache *ResourceCache) *types.Resource {
	cmID := types.BuildID("ConfigMap", cm.Namespace, cm.Name)

	resource := &types.Resource{
		ID:        cmID,
		Type:      "ConfigMap",
		Name:      cm.Name,
		Namespace: cm.Namespace,

		Status: types.ResourceStatus{
			Phase:   "Active",
			Ready:   "",
			Message: "",
		},

		Health: types.HealthHealthy,

		Relationships: types.Relationships{
			OwnedBy: ExtractOwners(cm),
			UsedBy:  FindReverseRelationships(cmID, types.RelDependsOn, cache),
		},

		Labels:      cm.Labels,
		Annotations: cm.Annotations,
		CreatedAt:   cm.CreationTimestamp.Time,
		Spec:        cm.Data,
		YAML:        marshalToYAML(cm),
	}

	return resource
}

// TransformSecret converts a Kubernetes Secret to our Resource model
func TransformSecret(secret *v1.Secret, cache *ResourceCache) *types.Resource {
	secretID := types.BuildID("Secret", secret.Namespace, secret.Name)

	resource := &types.Resource{
		ID:        secretID,
		Type:      "Secret",
		Name:      secret.Name,
		Namespace: secret.Namespace,

		Status: types.ResourceStatus{
			Phase:   "Active",
			Ready:   "",
			Message: "",
		},

		Health: types.HealthHealthy,

		Relationships: types.Relationships{
			OwnedBy: ExtractOwners(secret),
			UsedBy:  FindReverseRelationships(secretID, types.RelDependsOn, cache),
		},

		Labels:      secret.Labels,
		Annotations: secret.Annotations,
		CreatedAt:   secret.CreationTimestamp.Time,
		// Don't include actual secret data in Spec
		Spec: map[string]interface{}{
			"type": string(secret.Type),
		},
		YAML: marshalToYAML(secret),
	}

	return resource
}

// Helper functions for computing Pod status and health

func getPodReadyStatus(pod *v1.Pod) string {
	readyContainers := 0
	totalContainers := len(pod.Spec.Containers)

	for _, status := range pod.Status.ContainerStatuses {
		if status.Ready {
			readyContainers++
		}
	}

	return fmt.Sprintf("%d/%d", readyContainers, totalContainers)
}

func getPodMessage(pod *v1.Pod) string {
	// Check for container issues
	for _, status := range pod.Status.ContainerStatuses {
		if status.State.Waiting != nil {
			return status.State.Waiting.Reason
		}
		if status.State.Terminated != nil {
			return status.State.Terminated.Reason
		}
	}

	// Check for pod conditions
	for _, condition := range pod.Status.Conditions {
		if condition.Status == v1.ConditionFalse && condition.Reason != "" {
			return condition.Reason
		}
	}

	return ""
}

func computePodHealth(pod *v1.Pod) types.HealthState {
	phase := pod.Status.Phase

	// Check for failed states
	if phase == v1.PodFailed {
		return types.HealthError
	}

	// Check for container crash loops or errors
	for _, status := range pod.Status.ContainerStatuses {
		if status.State.Waiting != nil {
			reason := status.State.Waiting.Reason
			if reason == "CrashLoopBackOff" || reason == "ImagePullBackOff" || reason == "ErrImagePull" {
				return types.HealthError
			}
		}
		if status.State.Terminated != nil && status.State.Terminated.ExitCode != 0 {
			return types.HealthError
		}
	}

	// Check if all containers are ready
	readyContainers := 0
	for _, status := range pod.Status.ContainerStatuses {
		if status.Ready {
			readyContainers++
		}
	}

	if phase == v1.PodRunning && readyContainers == len(pod.Spec.Containers) {
		return types.HealthHealthy
	}

	if phase == v1.PodPending {
		return types.HealthWarning
	}

	return types.HealthUnknown
}

// Helper functions for Deployment

func getDeploymentPhase(deployment *appsv1.Deployment) string {
	if deployment.Status.ReadyReplicas == deployment.Status.Replicas {
		return "Available"
	}
	return "Progressing"
}

func getDeploymentMessage(deployment *appsv1.Deployment) string {
	if deployment.Status.ReadyReplicas < deployment.Status.Replicas {
		unavailable := deployment.Status.Replicas - deployment.Status.ReadyReplicas
		return fmt.Sprintf("%d replicas unavailable", unavailable)
	}
	return ""
}

func computeDeploymentHealth(deployment *appsv1.Deployment) types.HealthState {
	if deployment.Status.ReadyReplicas == 0 {
		return types.HealthError
	}
	if deployment.Status.ReadyReplicas < deployment.Status.Replicas {
		return types.HealthWarning
	}
	return types.HealthHealthy
}

// Helper functions for ReplicaSet

func computeReplicaSetHealth(rs *appsv1.ReplicaSet) types.HealthState {
	if rs.Status.ReadyReplicas == 0 && rs.Status.Replicas > 0 {
		return types.HealthError
	}
	if rs.Status.ReadyReplicas < rs.Status.Replicas {
		return types.HealthWarning
	}
	return types.HealthHealthy
}

// marshalToYAML converts a Kubernetes object to YAML string
func marshalToYAML(obj interface{}) string {
	data, err := yaml.Marshal(obj)
	if err != nil {
		return ""
	}
	return string(data)
}

// TransformNode converts a Kubernetes Node to our Resource model
func TransformNode(node *v1.Node, cache *ResourceCache) *types.Resource {
	nodeID := types.BuildID("Node", "", node.Name) // Nodes are cluster-scoped (no namespace)

	resource := &types.Resource{
		ID:        nodeID,
		Type:      "Node",
		Name:      node.Name,
		Namespace: "", // Nodes are cluster-scoped

		Status: types.ResourceStatus{
			Phase:   getNodePhase(node),
			Ready:   getNodeReadyStatus(node),
			Message: getNodeMessage(node),
		},

		Health: computeNodeHealth(node),

		Relationships: types.Relationships{
			Schedules: FindReverseRelationships(nodeID, types.RelScheduledOn, cache),
		},

		Labels:      node.Labels,
		Annotations: node.Annotations,
		CreatedAt:   node.CreationTimestamp.Time,
		Spec:        extractNodeSpec(node),
		YAML:        marshalToYAML(node),
	}

	return resource
}

// getNodePhase returns the node status phase
func getNodePhase(node *v1.Node) string {
	if node.Spec.Unschedulable {
		return "Unschedulable"
	}
	for _, condition := range node.Status.Conditions {
		if condition.Type == v1.NodeReady {
			if condition.Status == v1.ConditionTrue {
				return "Ready"
			}
			return "NotReady"
		}
	}
	return "Unknown"
}

// getNodeReadyStatus returns a human-readable ready status
func getNodeReadyStatus(node *v1.Node) string {
	for _, condition := range node.Status.Conditions {
		if condition.Type == v1.NodeReady {
			return string(condition.Status)
		}
	}
	return "Unknown"
}

// getNodeMessage returns condition messages for non-ready states
func getNodeMessage(node *v1.Node) string {
	var messages []string

	for _, condition := range node.Status.Conditions {
		if condition.Type == v1.NodeReady && condition.Status != v1.ConditionTrue {
			messages = append(messages, condition.Message)
		}
		// Check for pressure conditions
		if (condition.Type == v1.NodeMemoryPressure ||
			condition.Type == v1.NodeDiskPressure ||
			condition.Type == v1.NodePIDPressure) && condition.Status == v1.ConditionTrue {
			messages = append(messages, string(condition.Type))
		}
	}

	if len(messages) > 0 {
		return strings.Join(messages, "; ")
	}
	return ""
}

// computeNodeHealth determines health state based on conditions
func computeNodeHealth(node *v1.Node) types.HealthState {
	if node.Spec.Unschedulable {
		return types.HealthWarning
	}

	ready := false
	hasPressure := false

	for _, condition := range node.Status.Conditions {
		if condition.Type == v1.NodeReady {
			ready = (condition.Status == v1.ConditionTrue)
		}
		if (condition.Type == v1.NodeMemoryPressure ||
			condition.Type == v1.NodeDiskPressure ||
			condition.Type == v1.NodePIDPressure) && condition.Status == v1.ConditionTrue {
			hasPressure = true
		}
	}

	if !ready {
		return types.HealthError
	}
	if hasPressure {
		return types.HealthWarning
	}
	return types.HealthHealthy
}

// extractNodeSpec extracts relevant node spec information for display
func extractNodeSpec(node *v1.Node) map[string]interface{} {
	return map[string]interface{}{
		"capacity": map[string]string{
			"cpu":    node.Status.Capacity.Cpu().String(),
			"memory": node.Status.Capacity.Memory().String(),
			"pods":   node.Status.Capacity.Pods().String(),
		},
		"allocatable": map[string]string{
			"cpu":    node.Status.Allocatable.Cpu().String(),
			"memory": node.Status.Allocatable.Memory().String(),
			"pods":   node.Status.Allocatable.Pods().String(),
		},
		"nodeInfo": map[string]string{
			"osImage":          node.Status.NodeInfo.OSImage,
			"kernelVersion":    node.Status.NodeInfo.KernelVersion,
			"kubeletVersion":   node.Status.NodeInfo.KubeletVersion,
			"containerRuntime": node.Status.NodeInfo.ContainerRuntimeVersion,
		},
		"unschedulable": node.Spec.Unschedulable,
	}
}
