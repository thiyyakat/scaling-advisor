package testutil

import (
	"embed"
	"fmt"
	commontypes "github.com/gardener/scaling-advisor/api/common/types"
	"github.com/gardener/scaling-advisor/service/api"
	"github.com/gardener/scaling-advisor/service/pricing"
)

//go:embed testdata/*
var testDataFS embed.FS

func LoadTestInstanceTypeInfoAccess() (access api.InstanceTypeInfoAccess, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("%w: %w", api.ErrLoadInstanceTypeInfo, err)
		}
	}()
	testData, err := testDataFS.ReadFile("testdata/instance_type_infos.json")
	if err != nil {
		return
	}
	return pricing.GetInstancePricingFromData(commontypes.AWSCloudProvider, testData)
}
