// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName={csa}

// ClusterScalingAdvice is the schema to define cluster scaling advice for a cluster.
type ClusterScalingAdvice struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Spec defines the specification of ClusterScalingAdvice.
	Spec ClusterScalingAdviceSpec `json:"spec"`
	// Status defines the status of ClusterScalingAdvice.
	Status ClusterScalingAdviceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterScalingAdviceList is a list of ClusterScalingAdvice.
type ClusterScalingAdviceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	// Items is a slice of ClusterScalingAdvice.
	Items []ClusterScalingAdvice `json:"items"`
}

// ClusterScalingAdviceSpec defines the desired state of ClusterScalingAdvice.
type ClusterScalingAdviceSpec struct {
	// ConstraintRef is a reference to the ClusterScalingConstraint that this advice is based on.
	ConstraintRef ConstraintReference `json:"constraintRef"`
	// ScaleOutPlan is the plan for scaling out across node pools.
	// +optional
	ScaleOutPlan *ScaleOutPlan `json:"scaleOutPlan"`
	// ScaleInPlan is the plan for scaling in across node pools.
	ScaleInPlan *ScaleInPlan `json:"scaleInPlan"`
}

// ClusterScalingAdviceStatus defines the observed state of ClusterScalingAdvice.
type ClusterScalingAdviceStatus struct {
	// Feedback represents the lifecycle manager's feedback on the scaling advice.
	Feedback ClusterScalingAdviceFeedback `json:"feedback,omitempty"`
}

// ConstraintReference is a reference to the ClusterScalingConstraint for which this advice is generated.
type ConstraintReference struct {
	// Name is the name of the ClusterScalingConstraint.
	Name string `json:"name"`
	// Namespace is the namespace of the ClusterScalingConstraint.
	Namespace string `json:"namespace"`
}

// ScaleOutPlan is the plan for scaling out a node pool.
type ScaleOutPlan struct {
	// Items is the slice of scaling-out advice for a node pool.
	Items []ScaleItem `json:"Items"`
}

// ScaleInPlan is the plan for scaling in a node pool and/or targeted set of nodes.
type ScaleInPlan struct {
	// Items is the slice of scaling-in advice for a node pool.
	Items []ScaleItem `json:"items"`
	// NodeNames is the list of node names to be removed.
	NodeNames []string `json:"nodeNames"`
}

// ScaleItem is the unit of scaling advice for a node pool.
type ScaleItem struct {
	// NodePoolName is the name of the node pool.
	NodePoolName string `json:"nodePoolName"`
	// NodeTemplateName is the name of the node template.
	NodeTemplateName string `json:"nodeTemplateName"`
	// AvailabilityZone is the availability zone of the node pool.
	AvailabilityZone string `json:"availabilityZone"`
	// Delta is the delta change in the number of nodes for the node pool for this NodeTemplateName.
	Delta int32 `json:"delta"`
	// DesiredReplicas is the desired number of replicas for the node pool for this NodeTemplateName.
	DesiredReplicas int32 `json:"desiredReplicas"`
}

// ClusterScalingAdviceFeedback provides scale-in and scale-out error feedback from the lifecycle manager.
// Scaling advisor can refine its future scaling advice based on this feedback.
type ClusterScalingAdviceFeedback struct {
	// ScaleOutErrorInfos is the list of scale-out errors for the scaling advice.
	ScaleOutErrorInfos []ScaleOutErrorInfo `json:"scaleOutErrorInfos,omitempty"`
	// ScaleInErrorInfo is the scale-in error information for the scaling advice.
	ScaleInErrorInfo ScaleInErrorInfo `json:"scaleInErrorInfo,omitempty"`
}

// ScalingErrorType defines the type of scaling error.
type ScalingErrorType string

const (
	// ErrorTypeResourceExhausted indicates that the lifecycle manager could not create the instance due to resource exhaustion for an instance type in an availability zone.
	ErrorTypeResourceExhausted ScalingErrorType = "ResourceExhaustedError"
	// ErrorTypeCreationTimeout indicates that the lifecycle manager could not create the instance within its configured timeout despite multiple attempts.
	ErrorTypeCreationTimeout ScalingErrorType = "CreationTimeoutError"
)

// ScaleOutErrorInfo is the backoff information for each instance type + zone.
type ScaleOutErrorInfo struct {
	// AvailabilityZone is the availability zone of the node pool.
	AvailabilityZone string `json:"availabilityZone"`
	// InstanceType is the instance type of the node pool.
	InstanceType string `json:"instanceType"`
	// FailCount is the number of nodes that have failed creation.
	FailCount int32            `json:"failCount"`
	ErrorType ScalingErrorType `json:"errorType"`
}

// ScaleInErrorInfo is the information about nodes that could not be deleted for scale-in.
type ScaleInErrorInfo struct {
	// NodeNames is the list of node names that could not be deleted for scaled in.
	NodeNames []string `json:"nodeNames"`
}
