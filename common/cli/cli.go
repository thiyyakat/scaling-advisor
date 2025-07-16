package cli

import (
	"errors"
	"fmt"
	"github.com/spf13/pflag"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"runtime/debug"
	"strings"
)

const (
	ExitSuccess = iota
	// ExitErrParseOpts is the exit code indicating that the CLI has exited due to error parsing options.
	ExitErrParseOpts
	// ExitErrStart is the exit code indicating that there was an error starting the application.
	ExitErrStart
	// ExitErrShutdown is the exit code indicating that the application could not shutdown cleanly.
	ExitErrShutdown = 254
)

var (
	ErrParseArgs  = errors.New("cannot parse cli args")
	ErrMissingOpt = errors.New("missing option")
	ErrInvalidOpt = errors.New("invalid option value")
)

// CommonOptions is the set of constants CLI options which can be embedded into program specific options.
type CommonOptions struct {
	// Version option if set to true prints out the version of the program.
	Version bool
	// Host name/IP address for this service. Use "0.0.0.0"  to bind to all interfaces.
	Host string
	// Port is the HTTP port on which this service listens for requests.
	Port int
	// KubeConfigPath is the path to master kube-config.
	KubeConfigPath string
	// ProfilingEnable indicates whether this service should register the standard pprof HTTP handlers: /debug/pprof/*
	ProfilingEnabled bool
}

// MapCommonFlags adds the constants flags to the passed FlagSet.
func MapCommonFlags(flagSet *pflag.FlagSet, opts *CommonOptions) {
	flagSet.BoolVarP(&opts.Version, "version", "v", false, "-version prints the version information and quits")
	flagSet.StringVarP(&opts.KubeConfigPath, clientcmd.RecommendedConfigPathFlag, "k", opts.KubeConfigPath, "path to master kubeconfig - fallback to KUBECONFIG env-var")
	flagSet.StringVarP(&opts.Host, "host", "H", "", "host name to bind this service. Use 0.0.0.0 for all interfaces")
	flagSet.IntVarP(&opts.Port, "port", "P", 0, "listen port for REST API")
	flagSet.BoolVarP(&opts.ProfilingEnabled, "pprof", "p", false, "enable pprof profiling")
}

func (c CommonOptions) Validate() error {
	var errs []error
	if len(strings.TrimSpace(c.KubeConfigPath)) == 0 {
		errs = append(errs, fmt.Errorf("%w: --kubeconfig/-k flag is required", ErrMissingOpt))
	}
	if c.Port <= 0 {
		errs = append(errs, fmt.Errorf("%w: --port must be greater than 0", ErrInvalidOpt))
	}
	return errors.Join(errs...)
}

// PrintVersion prints the version from build information for the program.
func PrintVersion(programName string) {
	info, ok := debug.ReadBuildInfo()
	if ok && info.Main.Version != "" {
		fmt.Printf("%s version: %s\n", programName, info.Main.Version)
	} else {
		fmt.Printf("%s: binary build info not embedded\n", programName)
	}
}

// HandleErrorAndExit gracefully handles errors before exiting the program.
func HandleErrorAndExit(err error) {
	if errors.Is(err, pflag.ErrHelp) {
		os.Exit(ExitSuccess)
	}
	_, _ = fmt.Fprintf(os.Stderr, "Err: %v\n", err)
	os.Exit(ExitErrParseOpts)
}
