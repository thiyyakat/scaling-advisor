package pricing

import (
	commontypes "github.com/gardener/scaling-advisor/api/common/types"
	"github.com/gardener/scaling-advisor/service/api"
)

var _ api.GetInstancePricing = GetInstancePricing

func GetInstancePricing(provider commontypes.CloudProvider, pricingDataPath string) (api.InstancePricing, error) {
	return &InstancePricing{}, nil
}

var _ api.InstancePricing = (*InstancePricing)(nil)

type InstancePricing struct {
	CloudProvider commontypes.CloudProvider
	//TODO put pricing data in-memory.
}

func (a InstancePricing) GetPrice(region, instanceType string) (float64, error) {
	//TODO implement me
	panic("implement me")
}
