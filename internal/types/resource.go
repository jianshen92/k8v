package types

import "time"

// RelationshipType represents a type of relationship between resources
type RelationshipType string

const (
	RelOwnedBy   RelationshipType = "OwnedBy"
	RelOwns      RelationshipType = "Owns"
	RelDependsOn RelationshipType = "DependsOn"
	RelUsedBy    RelationshipType = "UsedBy"
	RelExposes   RelationshipType = "Exposes"
	RelExposedBy RelationshipType = "ExposedBy"
	RelRoutesTo  RelationshipType = "RoutesTo"
	RelRoutedBy  RelationshipType = "RoutedBy"
)

// GetReverseRelationshipType returns the reverse of a relationship type
func GetReverseRelationshipType(relType RelationshipType) RelationshipType {
	pairs := map[RelationshipType]RelationshipType{
		RelOwnedBy:   RelOwns,
		RelOwns:      RelOwnedBy,
		RelDependsOn: RelUsedBy,
		RelUsedBy:    RelDependsOn,
		RelExposes:   RelExposedBy,
		RelExposedBy: RelExposes,
		RelRoutesTo:  RelRoutedBy,
		RelRoutedBy:  RelRoutesTo,
	}
	return pairs[relType]
}

// Resource represents any Kubernetes resource with computed relationships
type Resource struct {
	// Identity
	ID        string `json:"id"`        // Unique: "type:namespace:name"
	Type      string `json:"type"`      // "Pod", "Deployment", "Service", etc.
	Name      string `json:"name"`
	Namespace string `json:"namespace"`

	// Status & Health
	Status ResourceStatus `json:"status"`
	Health HealthState    `json:"health"` // "healthy", "warning", "error", "unknown"

	// Relationships (the key part!)
	Relationships Relationships `json:"relationships"`

	// Metadata
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	CreatedAt   time.Time         `json:"createdAt"`

	// Raw data for detail views
	Spec interface{} `json:"spec,omitempty"` // Type-specific data
	YAML string      `json:"yaml"`           // Full YAML for viewing
}

// Relationships captures all connections between resources
type Relationships struct {
	// Ownership hierarchy
	OwnedBy []ResourceRef `json:"ownedBy"` // e.g., ReplicaSet owned by Deployment
	Owns    []ResourceRef `json:"owns"`    // e.g., Deployment owns ReplicaSets

	// Dependencies
	DependsOn []ResourceRef `json:"dependsOn"` // e.g., Pod depends on ConfigMap/Secret
	UsedBy    []ResourceRef `json:"usedBy"`    // e.g., ConfigMap used by Pods

	// Network relationships
	Exposes   []ResourceRef `json:"exposes"`   // e.g., Service exposes Pods
	ExposedBy []ResourceRef `json:"exposedBy"` // e.g., Pod exposed by Service
	RoutesTo  []ResourceRef `json:"routesTo"`  // e.g., Ingress routes to Service
	RoutedBy  []ResourceRef `json:"routedBy"`  // e.g., Service routed by Ingress
}

// ResourceRef is a lightweight reference to another resource
type ResourceRef struct {
	ID        string `json:"id"`        // "type:namespace:name"
	Type      string `json:"type"`      // "Pod", "Service", etc.
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// ResourceStatus contains type-specific status information
type ResourceStatus struct {
	Phase   string `json:"phase"`   // Type-specific: "Running", "Pending", "Active", etc.
	Ready   string `json:"ready"`   // e.g., "3/3" for Deployment replicas
	Message string `json:"message"` // Human-readable status explanation
}

// HealthState represents the high-level health indicator for visual representation
type HealthState string

const (
	HealthHealthy HealthState = "healthy" // Green: All good
	HealthWarning HealthState = "warning" // Yellow: Degraded or attention needed
	HealthError   HealthState = "error"   // Red: Failed or critical issue
	HealthUnknown HealthState = "unknown" // Gray: Cannot determine health
)

// BuildID creates a resource ID following the pattern "type:namespace:name"
func BuildID(resourceType, namespace, name string) string {
	if namespace == "" {
		// For cluster-scoped resources
		return resourceType + "::" + name
	}
	return resourceType + ":" + namespace + ":" + name
}

// NewResourceRef creates a ResourceRef from components
func NewResourceRef(resourceType, namespace, name string) ResourceRef {
	return ResourceRef{
		ID:        BuildID(resourceType, namespace, name),
		Type:      resourceType,
		Name:      name,
		Namespace: namespace,
	}
}

// GetRelationship returns the specified relationship field from a Resource
func (r *Resource) GetRelationship(relType RelationshipType) []ResourceRef {
	switch relType {
	case RelOwnedBy:
		return r.Relationships.OwnedBy
	case RelOwns:
		return r.Relationships.Owns
	case RelDependsOn:
		return r.Relationships.DependsOn
	case RelUsedBy:
		return r.Relationships.UsedBy
	case RelExposes:
		return r.Relationships.Exposes
	case RelExposedBy:
		return r.Relationships.ExposedBy
	case RelRoutesTo:
		return r.Relationships.RoutesTo
	case RelRoutedBy:
		return r.Relationships.RoutedBy
	default:
		return nil
	}
}
