package simulation

import (
	"context"
	"fmt"
	"github.com/gardener/scaling-advisor/service/api"
	"github.com/go-logr/logr"
	"golang.org/x/sync/errgroup"
	corev1 "k8s.io/api/core/v1"
)

// PartitionKey
type PartitionKey struct {
	NodePoolPriority     int
	NodeTemplatePriority int
}

type PartitionSimulationResult struct {
	AllNodeScores    []api.NodeScore
	WinnerScoreIndex int
	WinnerNode       *corev1.Node
}

func (s *PartitionSimulationResult) GetWinnerNodeScore() *api.NodeScore {
	if s.WinnerScoreIndex < 0 {
		return nil
	}
	return &s.AllNodeScores[s.WinnerScoreIndex]
}

// Partition represents a group of Simulations at the same priority level. We attempt to run simulations for the
// give partition and get a preferred NodeScore across Simulations of a partition before moving to Partition at the
// subsequent priority.
//
//		Example:1
//		np-a: 1 {nt-a: 1, nt-b: 2, nt-c: 1}
//		np-b: 2 {nt-q: 2, nt-r: 1, nt-s: 1}
//
//		p1: {PoolPriority: 1, NTPriority: 1, nt-a, nt-c}
//		p2: {PoolPriority: 1, NTPriority: 2, nt-b}
//	 p3: {PoolPriority: 2, NTPriority: 1, nt-r, nt-s}
//	 p4: {PoolPriority: 2, NTPriority: 2, nt-q}
//
//	Example:2
//		np-a: 1 {nt-a: 1, nt-b: 2, nt-c: 1}
//		np-b: 2 {nt-q: 2, nt-r: 1, nt-s: 1}
//		np-c: 1 {nt-x: 2, nt-y: 1}
//
//		p1: {PoolPriority: 1, NTPriority: 1, nt-a, nt-c, nt-y}
//		p2: {PoolPriority: 1, NTPriority: 2, nt-b, nt-x}
//		p3: {PoolPriority: 2, NTPriority: 1, nt-r, nt-s}
//		p4: {PoolPriority: 2, NTPriority: 2, nt-q}
type Partition struct {
	Key         PartitionKey
	Simulations []*Simulation
	Result      *PartitionSimulationResult
}

func (p *Partition) GetUnscheduledPodsOfWinner() []api.PodResourceInfo {
	if p.Result == nil || p.Result.GetWinnerNodeScore() == nil {
		return nil
	}
	return p.Result.GetWinnerNodeScore().UnscheduledPods
}

func (p *Partition) GetScaledNodeOfWinner() (*corev1.Node, error) {
	if p.Result == nil || p.Result.GetWinnerNodeScore() == nil {
		return nil, nil
	}
	winnerSimName := p.Result.GetWinnerNodeScore().SimulationName
	for _, s := range p.Simulations {
		if s.Args.Name == winnerSimName {
			return s.simNode, nil
		}
	}
	return nil, fmt.Errorf("no Node for winner simulation %q", winnerSimName)
}

func (p *Partition) Simulate(ctx context.Context, scorer api.NodeScorer, selector api.NodeScoreSelector) error {
	passCtx, passCancelFn := context.WithCancelCause(ctx)
	eg, passCtx := errgroup.WithContext(passCtx)
	log := logr.FromContextOrDiscard(ctx)
	// TODO: Create pool of similar NodeTemplate + Zone targets to scale and randomize over it so that we can have a balanced allocation across AZ.
	for _, sim := range p.Simulations {
		if err := sim.Start(passCtx, eg); err != nil {
			passCancelFn(err)
			return err
		}
	}
	if err := eg.Wait(); err != nil {
		// passCancelFn need not be invoked since ErrGroup takes care of it.
		return err
	}

	var scores []api.NodeScore

	for _, sim := range p.Simulations {
		assignments, err := sim.GetNodeAssignments()
		if err != nil {
			return fmt.Errorf("cannot get assignments for simulation %q: %w", sim.Args.Name, err)
		}
		if len(assignments) == 0 {
			log.Info("no pod assigments for simulation", "simulationName", sim.Args.Name)
			continue
		}
		nodeScoreArgs := api.NodeScoreArgs{
			SimulationName:   sim.Args.Name,
			Placement:        sim.GetScaledNodePlacementInfo(),
			ScaledAssignment: sim.GetScaledNodeAssignment(),
			Assignments:      assignments,
			UnscheduledPods:  sim.GetUnscheduledPods(),
		}
		score, err := scorer.Compute(nodeScoreArgs)
		if err != nil {
			return err
		}
		scores = append(scores, score)
	}

	winnerScoreIndex := selector(scores...)
	winnerNode, err := p.GetScaledNodeOfWinner()
	if err != nil {
		return err
	}

	p.Result = &PartitionSimulationResult{
		AllNodeScores:    scores,
		WinnerScoreIndex: winnerScoreIndex,
		WinnerNode:       winnerNode,
	}

	return nil
}
