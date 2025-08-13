package nodeutil

import corev1 "k8s.io/api/core/v1"

func GetInstanceType(node *corev1.Node) string {
	return node.Labels[corev1.LabelInstanceTypeStable]
}
