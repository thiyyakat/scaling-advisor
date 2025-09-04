package pricing

import (
	"encoding/json"
	"fmt"
	commontypes "github.com/gardener/scaling-advisor/api/common/types"
	"github.com/gardener/scaling-advisor/service/api"
	"os"
)

var _ api.GetProviderInstanceTypeInfoAccessFunc = GetInstancePricing

func GetInstancePricing(provider commontypes.CloudProvider, pricingDataPath string) (api.InstanceTypeInfoAccess, error) {
	data, err := os.ReadFile(pricingDataPath)
	if err != nil {
		return nil, err
	}
	return GetInstancePricingFromData(provider, data)
}

func GetInstancePricingFromData(provider commontypes.CloudProvider, data []byte) (api.InstanceTypeInfoAccess, error) {
	var ip infoAccess
	var err error
	ip.CloudProvider = provider
	ip.infosByPriceKey, err = parseInstanceTypeInfos(data)
	return &ip, err
}

func parseInstanceTypeInfos(data []byte) (map[api.PriceKey]api.InstanceTypeInfo, error) {
	var jsonEntries []api.InstanceTypeInfo
	//consider using streaming decoder instead
	err := json.Unmarshal(data, &jsonEntries)
	if err != nil {
		return nil, err
	}
	infosByPriceKey := make(map[api.PriceKey]api.InstanceTypeInfo, len(jsonEntries))
	for _, info := range jsonEntries {
		key := api.PriceKey{
			Name:   info.Name,
			Region: info.Region,
		}
		infosByPriceKey[key] = info
	}
	return infosByPriceKey, nil
}

var _ api.InstanceTypeInfoAccess = (*infoAccess)(nil)

type infoAccess struct {
	CloudProvider   commontypes.CloudProvider
	infosByPriceKey map[api.PriceKey]api.InstanceTypeInfo
}

func (a infoAccess) GetInfo(region, instanceType string) (info api.InstanceTypeInfo, err error) {
	info, ok := a.infosByPriceKey[api.PriceKey{
		Name:   instanceType,
		Region: region,
	}]
	if ok {
		return
	}
	err = fmt.Errorf("no instance type info found for instanceType %q in region %q ", instanceType, region)
	return
}
