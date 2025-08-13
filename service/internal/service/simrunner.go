package service

import (
	"context"
	"fmt"
	sacorev1alpha1 "github.com/gardener/scaling-advisor/api/core/v1alpha1"
	mkapi "github.com/gardener/scaling-advisor/minkapi/api"
	mkcore "github.com/gardener/scaling-advisor/minkapi/core"
	"github.com/gardener/scaling-advisor/service/api"
	"github.com/gardener/scaling-advisor/service/internal/service/simulation"
	"maps"
	"slices"
)

type simulationRunner struct {
	ctx               context.Context
	request           api.ScalingAdviceRequest
	minKAPIConfig     mkapi.MinKAPIConfig
	minKAPIServer     mkapi.Server
	schedulerLauncher api.SchedulerLauncher
	scorer            api.NodeScorer
	selector          api.NodeScoreSelector
	responseFn        api.ScalingAdviceResponseFn
}

func (r *simulationRunner) initMinKAPIServer() error {
	if r.minKAPIServer != nil {
		return nil
	}
	server, err := mkcore.NewInMemoryMinKAPI(r.ctx, r.minKAPIConfig)
	if err != nil {
		return fmt.Errorf("%w: %w", api.ErrInitFailed, err)
	}
	if err = server.Start(); err != nil {
		return fmt.Errorf("%w: %w", api.ErrInitFailed, err)
	}
	return nil
}

func (r *simulationRunner) populateBaseView() error {
	// populates the base View with the r.request.ClusterSnapshot
	// TODO implement me
	return nil
}

func (r *simulationRunner) Run() (err error) {
	defer func() {
		err = wrapError(r.request.ID, r.request.CorrelationID, err)
	}()
	// initialize MinKAPI and populate the base View.
	if err = r.initMinKAPIServer(); err != nil {
		return
	}
	if err = r.populateBaseView(); err != nil {
		return
	}
	var (
		winnerNodeScores []api.NodeScore
	)

	// Create partitions based on priorities that are defined at the NodePool and NodeTemplate level.
	partitions, err := r.createPartitions()
	if err != nil {
		return
	}

	// Run simulation for each such partition.
	for _, partition := range partitions {
		if err = partition.Simulate(r.ctx, r.scorer, r.selector); err != nil {
			return
		}
		unscheduledPods := partition.GetUnscheduledPodsOfWinner()
		if len(unscheduledPods) == 0 {
			break
		}
	}

	// If there is no scaling advice then return an error indicating the same.
	if len(winnerNodeScores) == 0 {
		err = api.ErrNoScalingAdvice
		return
	}

	// TODO create the simulation advice and then call the response function

	return nil
}

func (r *simulationRunner) createPartitions() ([]simulation.Partition, error) {
	var (
		allSimulations []*simulation.Simulation
		counter        int
	)
	for _, nodePool := range r.request.Constraint.Spec.NodePools {
		for _, nodeTemplate := range nodePool.NodeTemplates {
			for _, zone := range nodePool.AvailabilityZones {
				simulationName := fmt.Sprintf("%s-%s-%s-%d", nodePool.Name, zone, nodeTemplate.Name, counter)
				sim, err := r.createSimulation(simulationName, nodePool, nodeTemplate.Name, zone)
				if err != nil {
					return nil, err
				}
				allSimulations = append(allSimulations, sim)
			}
		}
	}

	partitionsByKey := make(map[simulation.PartitionKey]simulation.Partition)
	for _, sim := range allSimulations {
		pk := sim.GetPartitionKey()
		p, ok := partitionsByKey[pk]
		if !ok {
			p = simulation.Partition{
				Key:         pk,
				Simulations: []*simulation.Simulation{sim},
			}
		} else {
			p.Simulations = append(p.Simulations, sim)
		}
		partitionsByKey[pk] = p
	}

	partitions := slices.Collect(maps.Values(partitionsByKey))
	simulation.SortPartitions(partitions)
	return partitions, nil
}

func (r *simulationRunner) createSimulation(simulationName string, nodePool sacorev1alpha1.NodePool, nodeTemplateName string, zone string) (*simulation.Simulation, error) {
	simView, err := r.minKAPIServer.GetSimulationView()
	if err != nil {
		return nil, err
	}
	simArgs := &simulation.Args{
		Name:              simulationName,
		AvailabilityZone:  zone,
		NodePool:          nodePool,
		NodeTemplateName:  nodeTemplateName,
		SchedulerLauncher: r.schedulerLauncher,
		View:              simView,
	}
	return simulation.New(simArgs), nil
}
