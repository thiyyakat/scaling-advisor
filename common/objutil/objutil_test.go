// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package objutil

import (
	"fmt"
	"github.com/gardener/scaling-advisor/common/testutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
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
func TestLoadYamlIntoCoreRuntimeObj(t *testing.T) {
	tests := map[string]struct {
		filePath string
		retErr   error
	}{
		"valid yaml":        {filePath: "../testdata/pod-a.yaml", retErr: nil},
		"corrupt yaml":      {filePath: "../testdata/corrupt-pod-a.yaml", retErr: fmt.Errorf("failed to unmarshal object")},
		"non-yaml file":     {filePath: "./objutil.go", retErr: fmt.Errorf("failed to unmarshal object")},
		"non-existent path": {filePath: "../testdata/bingo.yaml", retErr: fmt.Errorf("failed to read")},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var pod1 corev1.Pod
			gotErr := LoadYamlIntoCoreRuntimeObj(tc.filePath, &pod1)
			testutil.AssertError(t, gotErr, tc.retErr)
		})
	}
}

func TestSetMetaObjectGVK(t *testing.T) {
	testPod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bingo",
			Namespace: "default",
		},
	}

	tests := map[string]struct {
		typeMeta    metav1.TypeMeta
		expectedGVK schema.GroupVersionKind
	}{
		"version and kind present": {
			typeMeta:    metav1.TypeMeta{APIVersion: "v1", Kind: "Pod"},
			expectedGVK: corev1.SchemeGroupVersion.WithKind("Pod"),
		},
		"kind absent": {
			typeMeta:    metav1.TypeMeta{APIVersion: "v1"},
			expectedGVK: schema.GroupVersionKind{Version: "v1"},
		},
		"version absent": {
			typeMeta:    metav1.TypeMeta{Kind: "Pod"},
			expectedGVK: schema.GroupVersionKind{Kind: "Pod"},
		},
		"version and kind absent": {
			typeMeta:    metav1.TypeMeta{},
			expectedGVK: corev1.SchemeGroupVersion.WithKind("Pod"),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			pod1 := testPod.DeepCopy()
			pod1.TypeMeta = tc.typeMeta
			obj1 := metav1.Object(pod1)

			SetMetaObjectGVK(obj1, corev1.SchemeGroupVersion.WithKind("Pod"))

			if rtObj1, ok := obj1.(runtime.Object); ok {
				if gotGVK := rtObj1.GetObjectKind().GroupVersionKind(); gotGVK != tc.expectedGVK {
					t.Errorf("GVK mismatch, got: %#v wanted: %#v", gotGVK, tc.expectedGVK)
				}
			}
		})
	}
}
