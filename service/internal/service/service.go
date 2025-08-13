package service

import (
	"context"
	"errors"
	"fmt"
	mkapi "github.com/gardener/scaling-advisor/minkapi/api"
	"github.com/gardener/scaling-advisor/service/api"
	"github.com/gardener/scaling-advisor/service/internal/scheduler"
	"github.com/go-logr/logr"
)

var _ api.ScalingAdvisorService = (*defaultScalingAdvisor)(nil)

type defaultScalingAdvisor struct {
	minKAPIConfig     mkapi.MinKAPIConfig
	minKAPIServer     mkapi.Server
	schedulerLauncher api.SchedulerLauncher
	scorer            api.NodeScorer
}

func New(config api.ScalingAdvisorServiceConfig, scorer api.NodeScorer) (api.ScalingAdvisorService, error) {
	schedulerLauncher, err := scheduler.NewLauncher(config.SchedulerConfigPath, config.MaxConcurrentSimulations)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", api.ErrInitFailed, err)
	}
	return &defaultScalingAdvisor{
		minKAPIConfig:     config.MinKAPIConfig,
		schedulerLauncher: schedulerLauncher,
		scorer:            scorer,
	}, nil
}

func (d *defaultScalingAdvisor) GenerateScalingAdvice(ctx context.Context, request api.ScalingAdviceRequest, responseFn api.ScalingAdviceResponseFn) (err error) {
	// wraps all errors with api.ErrGenScalingAdvice before return the error if any.
	defer func() {
		err = wrapError(request.ID, request.CorrelationID, err)
	}()
	unscheduledPods := getPodResourceInfos(request.Snapshot.GetUnscheduledPods())
	if len(unscheduledPods) == 0 {
		err = api.ErrNoUnscheduledPods
		return
	}

	genCtx := logr.NewContext(ctx, logr.FromContextOrDiscard(ctx).WithValues("requestID", request.ID, "correlationID", request.CorrelationID))
	simRunner := &simulationRunner{
		ctx:     genCtx,
		request: request,
		scorer:  d.scorer,
	}
	return simRunner.Run()
}

func wrapError(id string, correlationID string, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(api.ErrGenScalingAdvice, err) {
		return err
	}
	return fmt.Errorf("%w: could not process request with ID %q, CorrelationID %q: %w", api.ErrGenScalingAdvice, id, correlationID, err)
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
