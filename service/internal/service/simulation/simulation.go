package simulation

import (
	"cmp"
	"context"
	sacorev1alpha1 "github.com/gardener/scaling-advisor/api/core/v1alpha1"
	"github.com/gardener/scaling-advisor/common/nodeutil"
	mkapi "github.com/gardener/scaling-advisor/minkapi/api"
	"github.com/gardener/scaling-advisor/minkapi/core/typeinfo"
	"github.com/gardener/scaling-advisor/service/api"
	"github.com/go-logr/logr"
	"golang.org/x/sync/errgroup"
	corev1 "k8s.io/api/core/v1"
	"maps"
	"slices"
)

type Simulation struct {
	Args            *Args
	schedulerHandle api.SchedulerHandle
	nodeTemplate    *sacorev1alpha1.NodeTemplate
	simNode         *corev1.Node
	partitionKey    PartitionKey
	result          *result
}

type Phase string

const (
	PhaseInProgress Phase = "InProgress"
	PhaseSuccessful       = "Successful"
	PhaseFailed           = "Failed"
)

type result struct {
	Phase           Phase
	Err             error
	SimulationNode  *corev1.Node
	UnscheduledPods []*api.PodResourceInfo
	ScheduledPods   map[string][]*api.PodResourceInfo
}

type Args struct {
	// TODO this should be directly under Simulation and not in Args
	Name              string
	AvailabilityZone  string
	NodeTemplateName  string
	NodePool          sacorev1alpha1.NodePool
	SchedulerLauncher api.SchedulerLauncher
	View              mkapi.View
}

func New(args *Args) *Simulation {
	var nodeTemplate *sacorev1alpha1.NodeTemplate
	for _, nt := range args.NodePool.NodeTemplates {
		if nt.Name == args.NodeTemplateName {
			nodeTemplate = &nt
			break
		}
	}
	return &Simulation{
		Args:         args,
		nodeTemplate: nodeTemplate,
		partitionKey: PartitionKey{
			NodePoolPriority:     int(args.NodePool.Priority),
			NodeTemplatePriority: int(nodeTemplate.Priority),
		},
	}
}

func (s *Simulation) Start(ctx context.Context, eg *errgroup.Group) error {
	s.simNode = s.buildSimulationNode()
	if err := s.Args.View.CreateObject(typeinfo.NodesDescriptor.GVK, s.simNode); err != nil {
		return err
	}
	simCtx := newSimulationContext(ctx, s.Args.Name)
	schedulerHandle, err := s.launchSchedulerForSimulation(simCtx, s.Args.View)
	if err != nil {
		return err
	}
	s.schedulerHandle = schedulerHandle
	eg.Go(func() error {
		return s.track(ctx)
	})
	return nil
}

func (s *Simulation) GetPartitionKey() PartitionKey {
	return s.partitionKey
}

func (s *Simulation) GetScaledNodePlacementInfo() api.NodePlacementInfo {
	return api.NodePlacementInfo{
		NodePoolName:     s.Args.NodePool.Name,
		NodeTemplateName: s.nodeTemplate.Name,
		InstanceType:     s.nodeTemplate.InstanceType,
		AvailabilityZone: s.Args.AvailabilityZone,
	}
}

func (s *Simulation) GetScaledNodeAssignment() *api.NodePodAssignment {
	if s.result == nil {
		return nil
	}
	return &api.NodePodAssignment{
		Node:          GetNodeResourceInfo(s.result.SimulationNode),
		ScheduledPods: s.result.ScheduledPods[s.result.SimulationNode.Name],
	}
}

func (s *Simulation) GetUnscheduledPods() []*api.PodResourceInfo {
	if s.result == nil {
		return nil
	}
	return s.result.UnscheduledPods
}

func (s *Simulation) GetNodeAssignments() ([]*api.NodePodAssignment, error) {
	if s.result == nil {
		return nil, nil
	}
	nodeNames := slices.Collect(maps.Keys(s.result.ScheduledPods))
	nodeNames = slices.DeleteFunc(nodeNames, func(nodeName string) bool {
		return nodeName == s.result.SimulationNode.Name
	})
	nodes, err := s.Args.View.ListNodes(nodeNames...)
	if err != nil {
		return nil, err
	}
	var assignments []*api.NodePodAssignment
	for _, node := range nodes {
		nodeResources := GetNodeResourceInfo(node)
		podResources := s.result.ScheduledPods[node.Name]
		assignments = append(assignments, &api.NodePodAssignment{
			Node:          nodeResources,
			ScheduledPods: podResources,
		})
	}
	return assignments, nil
}

func (s *Simulation) launchSchedulerForSimulation(ctx context.Context, simView mkapi.View) (api.SchedulerHandle, error) {
	client, dynClient := simView.GetClients()
	informerFactory, dynInformerFactory := simView.GetInformerFactories()
	schedLaunchParams := &api.SchedulerLaunchParams{
		Client:             client,
		DynClient:          dynClient,
		InformerFactory:    informerFactory,
		DynInformerFactory: dynInformerFactory,
		EventSink:          simView.GetEventSink(),
	}
	return s.Args.SchedulerLauncher.Launch(ctx, schedLaunchParams)
}

func (s *Simulation) buildSimulationNode() *corev1.Node {
	/*
		create a simulation node based on the provided template, region, zone, labels, and taints.
		Add apiconstants.LabelSimulationID with the value of simulationName to the labels.
	*/
	return &corev1.Node{}
}

// track monitors the EventSink for scheduling events for all the unscheduled pods for a simulation run.
func (s *Simulation) track(ctx context.Context) error {
	/*
			NOTE: If there is an error then you return the error.
			If the ctx.Done then return the ctx.Err
			Get all the unscheduled pods from the simulation view.
			This function starts a loop which does the following till one of the following conditions is met:
		      1. All the pods are scheduled within the stabilization period OR
		      2. Stabilization period is over and there are still unscheduled pods.

			At the end of the loop it does the following:
			1. Updates the Simulation with unscheduled and scheduled pods.
			2. Signals the wait group by calling wg.Done to indicate that the simulation run is complete.
	*/
	return nil
}

func newSimulationContext(ctx context.Context, simulationName string) context.Context {
	log := logr.FromContextOrDiscard(ctx)
	return logr.NewContext(ctx, log.WithValues("simulationName", simulationName))
}

func SortPartitions(partitions []Partition) {
	slices.SortFunc(partitions, func(a, b Partition) int {
		npPriorityCmp := cmp.Compare(a.Key.NodePoolPriority, b.Key.NodePoolPriority)
		if npPriorityCmp != 0 {
			return npPriorityCmp
		}
		return cmp.Compare(a.Key.NodeTemplatePriority, b.Key.NodeTemplatePriority)
	})
}

func GetNodeResourceInfo(node *corev1.Node) *api.NodeResourceInfo {
	instanceType := nodeutil.GetInstanceType(node)
	return &api.NodeResourceInfo{
		Name:         node.Name,
		InstanceType: instanceType,
		Capacity:     node.Status.Capacity,
		Allocatable:  node.Status.Allocatable,
	}
}
