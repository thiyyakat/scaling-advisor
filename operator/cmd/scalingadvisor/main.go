package main

import (
	"github.com/gardener/scaling-advisor/api/common/constants"
	commoncli "github.com/gardener/scaling-advisor/common/cli"
	"github.com/gardener/scaling-advisor/operator/cmd/scalingadvisor/cli"
	"k8s.io/klog/v2"
	"os"
)

func main() {
	// Set up logr with klog backend using NewKlogr
	log := klog.NewKlogr()

	launchOpts, err := cli.ParseLaunchOptions(os.Args[1:])
	if err != nil {
		commoncli.HandleErrorAndExit(err)
	}
	if err = launchOpts.Validate(); err != nil {
		commoncli.HandleErrorAndExit(err)
	}
	if launchOpts.Version {
		commoncli.PrintVersion(constants.OperatorName)
		os.Exit(commoncli.ExitSuccess)
	}
	operatorConfig, err := launchOpts.LoadOperatorConfig()
	if err != nil {
		commoncli.HandleErrorAndExit(err)
	}

	log.Info("loaded configuration", "operatorConfig", operatorConfig)
}
