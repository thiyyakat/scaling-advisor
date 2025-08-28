package view

import (
	"context"
	"fmt"
	commontypes "github.com/gardener/scaling-advisor/api/common/types"
	"github.com/gardener/scaling-advisor/common/clientutil"
	"strings"
	"sync"
	"time"

	"github.com/gardener/scaling-advisor/common/objutil"
	"github.com/gardener/scaling-advisor/minkapi/server/podutil"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"

	"github.com/gardener/scaling-advisor/minkapi/api"
	"github.com/gardener/scaling-advisor/minkapi/server/eventsink"
	"github.com/gardener/scaling-advisor/minkapi/server/store"
	"github.com/gardener/scaling-advisor/minkapi/server/typeinfo"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	eventsv1 "k8s.io/api/events/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/uuid"
)

var (
	_ api.View                  = (*baseView)(nil)
	_ api.CreateSandboxViewFunc = NewSandboxView
)

type baseView struct {
	log       logr.Logger
	args      *api.ViewArgs
	mu        sync.RWMutex
	stores    map[schema.GroupVersionKind]*store.InMemResourceStore
	eventSink api.EventSink
}

func NewSandboxView(log logr.Logger, baseView api.View, args *api.ViewArgs) (api.View, error) {
	return NewSandbox(log, baseView, args)
}

func New(log logr.Logger, args *api.ViewArgs) (api.View, error) {
	stores := map[schema.GroupVersionKind]*store.InMemResourceStore{}
	for _, d := range typeinfo.SupportedDescriptors {
		stores[d.GVK] = store.NewInMemResourceStore(log, &api.ResourceStoreArgs{
			Name:          d.GVR.Resource,
			ObjectGVK:     d.GVK,
			ObjectListGVK: d.ListGVK,
			Scheme:        typeinfo.SupportedScheme,
			WatchConfig:   args.WatchConfig,
		})
		//stores[d.GVK] = store.NewInMemResourceStore(d.GVK, d.ListGVK, d.GVR.GroupResource().Resource, args.WatchConfig.QueueSize, args.WatchConfig.Timeout, typeinfo.SupportedScheme, log)
	}
	eventSink := eventsink.New(log)
	return &baseView{
		log:       log,
		args:      args,
		stores:    stores,
		eventSink: eventSink,
	}, nil
}

func (b *baseView) GetName() string {
	return b.args.Name
}

func (b *baseView) GetType() api.ViewType {
	return api.BaseViewType
}

func (b *baseView) GetClientFacades(clientType api.ClientType) (clientFacades commontypes.ClientFacades, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("%w: %w", api.ErrClientFacadesFailed, err)
		}
	}()
	if clientType == api.NetworkClient {
		clientFacades, err = clientutil.CreateNetworkClientFacades(b.log, b.args.KubeConfigPath, b.args.WatchConfig.Timeout)
		return
	} else {
		panic("inmem client type to be implemented")
	}
	return
}

func (b *baseView) GetEventSink() api.EventSink {
	return b.eventSink
}

func (b *baseView) GetResourceStore(gvk schema.GroupVersionKind) (api.ResourceStore, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	s, exists := b.stores[gvk]
	if !exists {
		return nil, fmt.Errorf("%w: store not found for GVK %q", api.ErrStoreNotFound, gvk)
	}
	return s, nil
}

func (b *baseView) GetObject(gvk schema.GroupVersionKind, fullName cache.ObjectName) (obj runtime.Object, err error) {
	s, err := b.GetResourceStore(gvk)
	if err != nil {
		return
	}
	key := fullName.String()
	obj, err = s.GetByKey(key)
	return
}

func (b *baseView) StoreObject(gvk schema.GroupVersionKind, obj metav1.Object) error {
	s, err := b.GetResourceStore(gvk)
	if err != nil {
		return err
	}

	name := obj.GetName()
	namePrefix := obj.GetGenerateName()
	if name == "" {
		if namePrefix == "" {
			return apierrors.NewBadRequest(fmt.Errorf("%w: missing both name and generateName in request for creating object of objGvk %q in %q namespace", api.ErrCreateObject, gvk, obj.GetNamespace()).Error())
		}
		name = typeinfo.GenerateName(namePrefix)
	}
	obj.SetName(name)

	createTimestamp := obj.GetCreationTimestamp()
	if (&createTimestamp).IsZero() { // only set creationTimestamp if not already set.
		obj.SetCreationTimestamp(metav1.Time{Time: time.Now()})
	}

	if obj.GetUID() == "" {
		obj.SetUID(uuid.NewUUID())
	}

	objutil.SetMetaObjectGVK(obj, gvk)

	err = s.Add(obj)
	if err != nil {
		return err
	}

	return nil
}

func (b *baseView) UpdateObject(gvk schema.GroupVersionKind, obj metav1.Object) error {
	s, err := b.GetResourceStore(gvk)
	if err != nil {
		return err
	}
	return s.Update(obj)
}

func (b *baseView) UpdatePodNodeBinding(podName cache.ObjectName, binding corev1.Binding) (*corev1.Pod, error) {
	gvk := typeinfo.PodsDescriptor.GVK
	obj, err := b.GetObject(gvk, podName)
	if err != nil {
		return nil, err
	}
	pod := obj.(*corev1.Pod)
	pod.Spec.NodeName = binding.Target.Name
	podutil.UpdatePodCondition(&pod.Status, &corev1.PodCondition{
		Type:   corev1.PodScheduled,
		Status: corev1.ConditionTrue,
	})
	err = b.UpdateObject(gvk, pod)
	if err != nil {
		return nil, err
	}
	return pod, nil
}

func (b *baseView) PatchObject(gvk schema.GroupVersionKind, objName cache.ObjectName, patchType types.PatchType, patchData []byte) (patchedObj runtime.Object, err error) {
	obj, err := b.GetObject(gvk, objName)
	if err != nil {
		return
	}
	err = objutil.PatchObject(obj, objName, patchType, patchData)
	if err != nil {
		err = fmt.Errorf("failed to patch object %q: %w", objName, err)
		return
	}
	mo, err := meta.Accessor(obj)
	if err != nil {
		err = fmt.Errorf("stored object with key %q is not metav1.Object: %w", objName, err)
		return
	}
	err = b.UpdateObject(gvk, mo)
	if err != nil {
		return
	}
	patchedObj = obj
	return
}

func (b *baseView) PatchObjectStatus(gvk schema.GroupVersionKind, objName cache.ObjectName, patchData []byte) (patchedObj runtime.Object, err error) {
	obj, err := b.GetObject(gvk, objName)
	if err != nil {
		return
	}
	err = objutil.PatchObjectStatus(obj, objName, patchData)
	if err != nil {
		err = fmt.Errorf("failed to patch object status of %q: %w", objName, err)
		return
	}
	mo, err := meta.Accessor(obj)
	if err != nil {
		err = fmt.Errorf("stored object with key %q is not metav1.Object: %w", objName, err)
		return
	}
	err = b.UpdateObject(gvk, mo)
	if err != nil {
		return
	}
	patchedObj = obj
	return
}

func (b *baseView) ListObjects(gvk schema.GroupVersionKind, criteria api.MatchCriteria) (runtime.Object, error) {
	s, err := b.GetResourceStore(gvk)
	if err != nil {
		return nil, err
	}
	listObj, err := s.List(criteria)
	if err != nil {
		return nil, err
	}
	return listObj, nil
}

func (b *baseView) WatchObjects(ctx context.Context, gvk schema.GroupVersionKind, startVersion int64, namespace string, labelSelector labels.Selector, eventCallback api.WatchEventCallback) error {
	s, err := b.GetResourceStore(gvk)
	if err != nil {
		return err
	}
	return s.Watch(ctx, startVersion, namespace, labelSelector, eventCallback)
}

func (b *baseView) DeleteObject(gvk schema.GroupVersionKind, fullName cache.ObjectName) error {
	s, err := b.GetResourceStore(gvk)
	if err != nil {
		return err
	}
	return s.Delete(fullName.String())
}

func (b *baseView) DeleteObjects(gvk schema.GroupVersionKind, criteria api.MatchCriteria) error {
	s, err := b.GetResourceStore(gvk)
	if err != nil {
		return err
	}
	return s.DeleteObjects(criteria)
}

func (b *baseView) ListNodes(matchingNodeNames ...string) ([]*corev1.Node, error) {
	nodeNamesSet := sets.New(matchingNodeNames...)
	c := api.MatchCriteria{
		Names: nodeNamesSet,
	}
	gvk := typeinfo.NodesDescriptor.GVK
	s, err := b.GetResourceStore(gvk)
	if err != nil {
		return nil, err
	}
	objs, err := s.ListMetaObjects(c)
	if err != nil {
		return nil, err
	}
	nodes := make([]*corev1.Node, 0, len(objs))
	for _, obj := range objs {
		nodes = append(nodes, obj.(*corev1.Node))
	}
	return nodes, nil
}

func (b *baseView) ListPods(namespace string, matchingPodNames ...string) ([]*corev1.Pod, error) {
	if len(strings.TrimSpace(namespace)) == 0 {
		return nil, apierrors.NewBadRequest("cannot list pods without namespace")
	}
	podNamesSet := sets.New(matchingPodNames...)
	c := api.MatchCriteria{
		Namespace: namespace,
		Names:     podNamesSet,
	}
	gvk := typeinfo.PodsDescriptor.GVK
	s, err := b.GetResourceStore(gvk)
	if err != nil {
		return nil, err
	}
	objs, err := s.ListMetaObjects(c)
	if err != nil {
		return nil, err
	}
	pods := make([]*corev1.Pod, 0, len(objs))
	for _, obj := range objs {
		pods = append(pods, obj.(*corev1.Pod))
	}
	return pods, nil
}

func (b *baseView) ListEvents(namespace string) ([]*eventsv1.Event, error) {
	if len(strings.TrimSpace(namespace)) == 0 {
		return nil, apierrors.NewBadRequest("cannot list events without namespace")
	}
	c := api.MatchCriteria{
		Namespace: namespace,
	}
	gvk := typeinfo.EventsDescriptor.GVK
	s, err := b.GetResourceStore(gvk)
	if err != nil {
		return nil, err
	}
	objs, err := s.ListMetaObjects(c)
	if err != nil {
		return nil, err
	}
	events := make([]*eventsv1.Event, 0, len(objs))
	for _, obj := range objs {
		events = append(events, obj.(*eventsv1.Event))
	}
	return events, nil
}

func (b *baseView) GetKubeConfigPath() string {
	return b.args.KubeConfigPath
}

func (b *baseView) Close() {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, s := range b.stores {
		s.Shutdown()
	}
}
