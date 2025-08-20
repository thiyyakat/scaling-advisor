// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package objutil

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"reflect"
	"testing"
)

func TestResourceListToMapInt64(t *testing.T) {
	tests := []struct {
		name string
		args corev1.ResourceList
		want map[corev1.ResourceName]int64
	}{
		{
			name: "simple-cpu_mem_ephemeral-storage",
			args: corev1.ResourceList{
				corev1.ResourceMemory:           *resource.NewQuantity(1024, resource.BinarySI),
				corev1.ResourceCPU:              *resource.NewQuantity(2, resource.DecimalSI),
				corev1.ResourceEphemeralStorage: resource.MustParse("100Mi"),
			},
			want: map[corev1.ResourceName]int64{
				corev1.ResourceMemory:           1024,
				corev1.ResourceCPU:              2,
				corev1.ResourceEphemeralStorage: 100 * 1024 * 1024,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ResourceListToMapInt64(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ResourceListToMapInt64() = %v, want %v", got, tt.want)
			}
		})
	}
}
