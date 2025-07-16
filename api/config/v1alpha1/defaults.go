package v1alpha1

import (
	commontypes "github.com/gardener/scaling-advisor/api/common/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

const (
	defaultLeaderElectionResourceLock = "leases"
	defaultLeaderElectionResourceName = "scaling-advisor-operator-leader-election"
	defaultHealthProbePort            = 2751
	defaultMetricsPort                = 2752
)

// SetDefaults_ClientConnectionConfiguration sets defaults for the k8s client connection.
func SetDefaults_ClientConnectionConfiguration(clientConnConfig *ClientConnectionConfiguration) {
	if clientConnConfig.QPS == 0.0 {
		clientConnConfig.QPS = 100.0
	}
	if clientConnConfig.Burst == 0 {
		clientConnConfig.Burst = 120
	}
}

// SetDefaults_LeaderElectionConfiguration sets defaults for the leader election of the scaling-advisor operator.
func SetDefaults_LeaderElectionConfiguration(leaderElectionConfig *LeaderElectionConfiguration) {
	zero := metav1.Duration{}
	if leaderElectionConfig.LeaseDuration == zero {
		leaderElectionConfig.LeaseDuration = metav1.Duration{Duration: 15 * time.Second}
	}
	if leaderElectionConfig.RenewDeadline == zero {
		leaderElectionConfig.RenewDeadline = metav1.Duration{Duration: 10 * time.Second}
	}
	if leaderElectionConfig.RetryPeriod == zero {
		leaderElectionConfig.RetryPeriod = metav1.Duration{Duration: 2 * time.Second}
	}
	if leaderElectionConfig.ResourceLock == "" {
		leaderElectionConfig.ResourceLock = defaultLeaderElectionResourceLock
	}
	if leaderElectionConfig.ResourceName == "" {
		leaderElectionConfig.ResourceName = defaultLeaderElectionResourceName
	}
}

func SetDefaults_HealthProbes(healthProbesConfig *commontypes.Server) {
	if healthProbesConfig.Port == 0 {
		healthProbesConfig.Port = defaultHealthProbePort
	}
}

func SetDefaults_Metrics(metricsConfig *commontypes.Server) {
	if metricsConfig.Port == 0 {
		metricsConfig.Port = defaultMetricsPort
	}
}
