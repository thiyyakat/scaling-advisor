package main

import (
	"github.com/gardener/scaling-advisor/api/common/constants"
	clicommon "github.com/gardener/scaling-advisor/common/cli"
)

func main() {
	clicommon.PrintVersion(constants.OperatorName)
}
