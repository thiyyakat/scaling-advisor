// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"errors"
	"flag"
	"fmt"
	commonconstants "github.com/gardener/scaling-advisor/api/common/constants"
	"github.com/gardener/scaling-advisor/api/minkapi"
	commoncli "github.com/gardener/scaling-advisor/common/cli"
	"github.com/spf13/pflag"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"os"
	"strings"
)

// MainOpts is a struct that encapsulates target fields for CLI options parsing.
type MainOpts struct {
	minkapi.Config
}

// ParseProgramFlags parses the command line arguments and returns MainOpts.
func ParseProgramFlags(args []string) (*MainOpts, error) {
	flagSet, mainOpts := setupFlagsToOpts()
	err := flagSet.Parse(args)
	if err != nil {
		return nil, err
	}
	err = validateMainOpts(mainOpts)
	if err != nil {
		return nil, err
	}
	return mainOpts, nil
}

func setupFlagsToOpts() (*pflag.FlagSet, *MainOpts) {
	var mainOpts MainOpts
	flagSet := pflag.NewFlagSet(minkapi.ProgramName, pflag.ContinueOnError)

	mainOpts.KubeConfigPath = os.Getenv(clientcmd.RecommendedConfigPathEnvVar)
	if mainOpts.KubeConfigPath == "" {
		mainOpts.KubeConfigPath = minkapi.DefaultKubeConfigPath
	}
	if mainOpts.Port == 0 {
		mainOpts.Port = commonconstants.DefaultMinKAPIPort
	}
	commoncli.MapServerConfigFlags(flagSet, &mainOpts.ServerConfig)
	flagSet.IntVarP(&mainOpts.WatchConfig.QueueSize, "watch-queue-size", "s", minkapi.DefaultWatchQueueSize, "max number of events to queue per watcher")
	flagSet.DurationVarP(&mainOpts.WatchConfig.Timeout, "watch-timeout", "t", minkapi.DefaultWatchTimeout, "watch timeout after which connection is closed and watch removed")
	flagSet.StringVarP(&mainOpts.BasePrefix, "base-prefix", "b", minkapi.DefaultBasePrefix, "base path prefix for the base view of the minkapi service")

	klogFlagSet := flag.NewFlagSet("klog", flag.ContinueOnError)
	klog.InitFlags(klogFlagSet)

	// Merge klog flags into pflag
	flagSet.AddGoFlagSet(klogFlagSet)

	return flagSet, &mainOpts
}

func validateMainOpts(opts *MainOpts) error {
	var errs []error
	errs = append(errs, commoncli.ValidateServerConfigFlags(opts.ServerConfig))
	if len(strings.TrimSpace(opts.KubeConfigPath)) == 0 {
		errs = append(errs, fmt.Errorf("%w: --kubeconfig/-k", minkapi.ErrMissingOpt))
	}
	return errors.Join(errs...)
}
