package service

import (
	"context"
	"fmt"
	mkapi "github.com/gardener/scaling-advisor/minkapi/api"
	mkcore "github.com/gardener/scaling-advisor/minkapi/server"
	"github.com/gardener/scaling-advisor/service/api"
	"github.com/gardener/scaling-advisor/service/internal/scheduler"
	"github.com/gardener/scaling-advisor/service/internal/service/generator"
	"github.com/gardener/scaling-advisor/service/internal/service/simulation"
	"github.com/go-logr/logr"
)

var _ api.ScalingAdvisorService = (*defaultScalingAdvisor)(nil)

type defaultScalingAdvisor struct {
	minKAPIConfig     mkapi.MinKAPIConfig
	minKAPIServer     mkapi.Server
	schedulerLauncher api.SchedulerLauncher
	pricer            api.InstanceTypeInfoAccess
	weighsFn          api.GetWeightsFunc
	scorer            api.NodeScorer
	selector          api.NodeScoreSelector
}

func New(config api.ScalingAdvisorServiceConfig,
	pricer api.InstanceTypeInfoAccess,
	weights api.GetWeightsFunc,
	scorer api.NodeScorer,
	selector api.NodeScoreSelector) (api.ScalingAdvisorService, error) {
	schedulerLauncher, err := scheduler.NewLauncher(config.SchedulerConfigPath, config.MaxConcurrentSimulations)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", api.ErrInitFailed, err)
	}
	return &defaultScalingAdvisor{
		minKAPIConfig:     config.MinKAPIConfig,
		schedulerLauncher: schedulerLauncher,
		pricer:            pricer,
		weighsFn:          weights,
		scorer:            scorer,
		selector:          selector,
	}, nil
}

func (d *defaultScalingAdvisor) Start(ctx context.Context) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("%w: %w", api.ErrInitFailed, err)
		}
	}()
	log := logr.FromContextOrDiscard(ctx)
	d.minKAPIServer, err = mkcore.NewDefaultInMemory(log, d.minKAPIConfig)
	if err != nil {
		return
	}
	if err = d.minKAPIServer.Start(ctx); err != nil {
		return
	}
	return nil
}

func (d *defaultScalingAdvisor) Stop(ctx context.Context) error {
	if d.minKAPIServer != nil {
		return d.minKAPIServer.Stop(ctx)
	}
	return nil
}

func (d *defaultScalingAdvisor) GenerateAdvice(ctx context.Context, request api.ScalingAdviceRequest) <-chan api.ScalingAdviceEvent {
	eventCh := make(chan api.ScalingAdviceEvent)
	go func() {
		unscheduledPods := getPodResourceInfos(request.Snapshot.GetUnscheduledPods())
		if len(unscheduledPods) == 0 {
			err := api.AsGenerateError(request.ID, request.CorrelationID, fmt.Errorf("%w: no unscheduled pods found", api.ErrNoUnscheduledPods))
			eventCh <- api.ScalingAdviceEvent{
				Err: err,
			}
			return
		}
		genCtx := logr.NewContext(ctx, logr.FromContextOrDiscard(ctx).WithValues("requestID", request.ID, "correlationID", request.CorrelationID))
		g := generator.New(genCtx, &generator.Args{
			Pricer:            d.pricer,
			WeightsFn:         d.weighsFn,
			Scorer:            d.scorer,
			Selector:          d.selector,
			CreateSimFn:       simulation.New,
			CreateSimGroupsFn: simulation.CreateSimulationGroups,
			Request:           request,
			EventChannel:      eventCh,
		})
		g.Generate()
	}()
	return eventCh
}

func getPodResourceInfos(podInfos []api.PodInfo) []api.PodResourceInfo {
	podResourceInfos := make([]api.PodResourceInfo, 0, len(podInfos))
	for _, podInfo := range podInfos {
		podResourceInfos = append(podResourceInfos, api.PodResourceInfo{
			UID:                podInfo.UID,
			NamespacedName:     podInfo.NamespacedName,
			AggregatedRequests: podInfo.AggregatedRequests,
		})
	}
	return podResourceInfos
}
