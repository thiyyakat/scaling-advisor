package cli

import (
	"fmt"
	"github.com/gardener/scaling-advisor/api/common/constants"
	"github.com/gardener/scaling-advisor/api/config/v1alpha1"
	commoncli "github.com/gardener/scaling-advisor/common/cli"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"os"
)

var ErrLoadOperatorConfig = fmt.Errorf("cannot load %q operator config", constants.OperatorName)

type LaunchOptions struct {
	ConfigFile string
	Version    bool
}

// ParseLaunchOptions parses the CLI arguments for the scaling-advisor operator.
func ParseLaunchOptions(cliArgs []string) (*LaunchOptions, error) {
	launchOpts := &LaunchOptions{}
	flagSet := pflag.NewFlagSet(constants.OperatorName, pflag.ContinueOnError)
	launchOpts.mapFlags(flagSet)
	err := flagSet.Parse(cliArgs)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", commoncli.ErrParseArgs, err)
	}
	return launchOpts, nil
}

// LoadOperatorConfig loads the operator configuration from the ConfigFile specified in the LaunchOptions.
func (o *LaunchOptions) LoadOperatorConfig() (*v1alpha1.OperatorConfiguration, error) {
	configScheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(configScheme); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrLoadOperatorConfig, err)
	}
	configDecoder := serializer.NewCodecFactory(configScheme).UniversalDecoder()
	configBytes, err := os.ReadFile(o.ConfigFile)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrLoadOperatorConfig, err)
	}
	operatorConfig := &v1alpha1.OperatorConfiguration{}
	if err := runtime.DecodeInto(configDecoder, configBytes, operatorConfig); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrLoadOperatorConfig, err)
	}
	return operatorConfig, nil
}

func (o *LaunchOptions) Validate() error {
	if len(o.ConfigFile) == 0 && !o.Version {
		return fmt.Errorf("%w: atleast version or config should be specified", commoncli.ErrMissingOpt)
	}
	return nil
}

func (o *LaunchOptions) mapFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.ConfigFile, "config", o.ConfigFile, "path to the config file")
	fs.BoolVarP(&o.Version, "version", "v", o.Version, "print version and exit")
}
