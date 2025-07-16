package v1alpha1

import (
	commontypes "github.com/gardener/scaling-advisor/api/common/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// OperatorConfiguration defines the configuration for the scaling-advisor operator.
type OperatorConfiguration struct {
	metav1.TypeMeta `json:",inline"`
	// ClientConnection defines the configuration for constructing a kube client.
	ClientConnection ClientConnectionConfiguration `json:"runtimeClientConnection"`
	// LeaderElection defines the configuration for leader election.
	LeaderElection LeaderElectionConfiguration `json:"leaderElection"`
	// HealthProbes is the configuration for serving the healthz and readyz endpoints.
	// +optional
	HealthProbes *commontypes.Server `json:"healthProbes,omitempty"`
	// Metrics is the configuration for serving the metrics endpoint.
	// +optional
	Metrics                   *commontypes.Server `json:"metrics,omitempty"`
	commontypes.CommonOptions `json:",inline"`
}

// ClientConnectionConfiguration defines the configuration for constructing a client.
type ClientConnectionConfiguration struct {
	// QPS controls the number of queries per second allowed for this connection.
	QPS float32 `json:"qps"`
	// Burst allows extra queries to accumulate when a client is exceeding its rate.
	Burst int `json:"burst"`
	// ContentType is the content type used when sending data to the server from this client.
	ContentType string `json:"contentType"`
	// AcceptContentTypes defines the Accept header sent by clients when connecting to the server,
	// overriding the default value of 'application/json'. This field will control all connections
	// to the server used by a particular client.
	AcceptContentTypes string `json:"acceptContentTypes"`
}

// LeaderElectionConfiguration defines the configuration for the leader election.
type LeaderElectionConfiguration struct {
	// Enabled specifies whether leader election is enabled. Set this
	// to true when running replicated instances of the operator for high availability.
	Enabled bool `json:"enabled"`
	// LeaseDuration is the duration that non-leader candidates will wait
	// after observing a leadership renewal until attempting to acquire
	// leadership of the occupied but un-renewed leader slot. This is effectively the
	// maximum duration that a leader can be stopped before it is replaced
	// by another candidate. This is only applicable if leader election is
	// enabled.
	LeaseDuration metav1.Duration `json:"leaseDuration"`
	// RenewDeadline is the interval between attempts by the acting leader to
	// renew its leadership before it stops leading. This must be less than or
	// equal to the lease duration.
	// This is only applicable if leader election is enabled.
	RenewDeadline metav1.Duration `json:"renewDeadline"`
	// RetryPeriod is the duration leader elector clients should wait
	// between attempting acquisition and renewal of leadership.
	// This is only applicable if leader election is enabled.
	RetryPeriod metav1.Duration `json:"retryPeriod"`
	// ResourceLock determines which resource lock to use for leader election.
	// This is only applicable if leader election is enabled.
	ResourceLock string `json:"resourceLock"`
	// ResourceName determines the name of the resource that leader election
	// will use for holding the leader lock.
	// This is only applicable if leader election is enabled.
	ResourceName string `json:"resourceName"`
	// ResourceNamespace determines the namespace in which the leader
	// election resource will be created.
	// This is only applicable if leader election is enabled.
	ResourceNamespace string `json:"resourceNamespace"`
}
