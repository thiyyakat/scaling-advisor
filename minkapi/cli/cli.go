// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/gardener/scaling-advisor/minkapi/api"
	"github.com/gardener/scaling-advisor/minkapi/server"
	"github.com/go-logr/logr"
	"os"
	"strings"
	"time"

	commonconstants "github.com/gardener/scaling-advisor/api/common/constants"
	commoncli "github.com/gardener/scaling-advisor/common/cli"
	"github.com/spf13/pflag"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

// MainOpts is a struct that encapsulates target fields for CLI options parsing.
type MainOpts struct {
	api.MinKAPIConfig
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
	flagSet := pflag.NewFlagSet(api.ProgramName, pflag.ContinueOnError)

	mainOpts.KubeConfigPath = os.Getenv(clientcmd.RecommendedConfigPathEnvVar)
	if mainOpts.KubeConfigPath == "" {
		mainOpts.KubeConfigPath = api.DefaultKubeConfigPath
	}
	if mainOpts.Port == 0 {
		mainOpts.Port = commonconstants.DefaultMinKAPIPort
	}
	commoncli.MapServerConfigFlags(flagSet, &mainOpts.ServerConfig)
	flagSet.IntVarP(&mainOpts.WatchConfig.QueueSize, "watch-queue-size", "s", api.DefaultWatchQueueSize, "max number of events to queue per watcher")
	flagSet.DurationVarP(&mainOpts.WatchConfig.Timeout, "watch-timeout", "t", api.DefaultWatchTimeout, "watch timeout after which connection is closed and watch removed")
	flagSet.StringVarP(&mainOpts.BasePrefix, "base-prefix", "b", api.DefaultBasePrefix, "base path prefix for the base view of the minkapi service")

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
		errs = append(errs, fmt.Errorf("%w: --kubeconfig/-k", api.ErrMissingOpt))
	}
	return errors.Join(errs...)
}

// App represents a service application and its top level application context and cancel func along with any exit code.
// Used by top level cli/launch code.
// TODO: consider moving this to commontypes
type App struct {
	Service api.Server
	Ctx     context.Context
	Cancel  context.CancelFunc
}

// LaunchApp is a helper function used to parse cli args, construct and start the MinKAPI server.
//
// On success, returns an initialized App which holds the minkapi Service, the App Context (which has been setup for SIGINT and SIGTERM cancellation and holds a logger),
// and the Cancel func which callers are expected to defer in their main routines.
//
// On error, it will log the error to standard error and return the exitCode that callers are expected to exit the process with.
func LaunchApp() (app App, exitCode int) {
	app.Ctx, app.Cancel = commoncli.CreateAppContext()
	log := logr.FromContextOrDiscard(app.Ctx)
	commoncli.PrintVersion(api.ProgramName)
	mainOpts, err := ParseProgramFlags(os.Args[1:])
	if err != nil {
		if errors.Is(err, pflag.ErrHelp) {
			return
		}
		_, _ = fmt.Fprintf(os.Stderr, "Err: %v\n", err)
		exitCode = commoncli.ExitErrParseOpts
		return
	}
	app.Service, err = server.NewDefaultInMemory(log, mainOpts.MinKAPIConfig)
	if err != nil {
		log.Error(err, "failed to initialize InMemoryKAPI")
		exitCode = commoncli.ExitErrStart
		return
	}
	// Begin the service in a goroutine
	go func() {
		if err := app.Service.Start(logr.NewContext(app.Ctx, log)); err != nil {
			if errors.Is(err, api.ErrStartFailed) {
				log.Error(err, "failed to start service")
			} else {
				log.Error(err, fmt.Sprintf("%s start failed", api.ProgramName))
			}
		}
	}()
	return
}

func ShutdownApp(app *App) (exitCode int) {
	// Create a context with a 5-second timeout for shutdown
	shutDownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	log := logr.FromContextOrDiscard(app.Ctx)

	// Perform shutdown
	if err := app.Service.Stop(shutDownCtx); err != nil {
		log.Error(err, fmt.Sprintf(" %s shutdown failed", api.ProgramName))
		exitCode = commoncli.ExitErrShutdown
		return
	}
	log.Info(fmt.Sprintf("%s shutdown gracefully.", api.ProgramName))
	exitCode = commoncli.ExitSuccess
	return
}
