package scorer

import (
	"fmt"
	commontypes "github.com/gardener/scaling-advisor/api/common/types"
	"github.com/gardener/scaling-advisor/service/api"
)

var _ api.NodeScoreSelector = GetNodeScoreSelector

func GetNodeScoreSelector(nodeScores ...api.NodeScore) int {
	//TODO implement me
	panic("implement me")
}

var _ api.GetNodeScorer = GetNodeScorer

func GetNodeScorer(scoringStrategy commontypes.NodeScoringStrategy, instancePricing api.InstancePricing) (api.NodeScorer, error) {
	switch scoringStrategy {
	case commontypes.LeastCostNodeScoringStrategy:
		return &LeastCost{instancePricing: instancePricing}, nil
	case commontypes.LeastWasteNodeScoringStrategy:
		return &LeastWaste{instancePricing: instancePricing}, nil
	default:
		return nil, fmt.Errorf("%w: unsupported %q", api.ErrUnsupportedNodeScoringStrategy, scoringStrategy)
	}
}

var _ api.NodeScorer = (*LeastCost)(nil)

type LeastCost struct {
	instancePricing api.InstancePricing
	// TODO
}

func (l LeastCost) Compute(args api.NodeScoreArgs) (api.NodeScore, error) {
	//TODO implement me
	panic("implement me")
}

var _ api.NodeScorer = (*LeastWaste)(nil)

type LeastWaste struct {
	instancePricing api.InstancePricing
	// TODO
}

func (l LeastWaste) Compute(args api.NodeScoreArgs) (api.NodeScore, error) {
	//TODO implement me
	panic("implement me")
}
