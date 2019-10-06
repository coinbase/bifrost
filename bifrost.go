package main

import (
	"fmt"
	"time"
	"os"

	"github.com/coinbase/bifrost/client"
	"github.com/coinbase/bifrost/aws"
	"github.com/coinbase/bifrost/deployer"
	"github.com/coinbase/step/utils/run"
	"github.com/coinbase/step/utils/to"

	"github.com/coinbase/bifrost/cadence/common"
	"github.com/coinbase/bifrost/cadence"
	"go.uber.org/cadence/worker"
	cadenceclient "go.uber.org/cadence/client"
	"go.uber.org/zap"
)

// This needs to be done as part of a bootstrap step when the process starts.
// The workers are supposed to be long running.
func startWorkers(h *common.SampleHelper) {
	// Configure worker options.
	workerOptions := worker.Options{
		MetricsScope: h.Scope,
		Logger:       h.Logger,
	}
	h.StartWorkers(h.Config.DomainName, cadence.ApplicationName, workerOptions)
}

func main() {
	var arg, command string
	switch len(os.Args) {
	case 1:
		fmt.Println("Starting Lambda")
		run.LambdaTasks(deployer.TaskHandlers())
	case 2:
		command = os.Args[1]
		arg = ""
	case 3:
		command = os.Args[1]
		arg = os.Args[2]
	default:
		printUsage() // Print how to use and exit
	}

	step_fn := to.Strp("coinbase-bifrost")

	var h common.SampleHelper
	h.SetupServiceConfig()

	switch command {
	case "json":
		// This is required to use the step to deploy
		run.JSON(deployer.StateMachine())
	case "deploy":
		// Send Configuration to the deployer
		// arg is a filename
		err := client.Deploy(step_fn, &arg)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	case "worker":
		// Run cadence worker
		startWorkers(&h)

		select {}
	case "trigger":
		release, err := client.DeployS3(&aws.ClientsStr{}, &arg)

		if err != nil {
			h.Logger.Fatal("failed constructing release", zap.Error(err))
		}

		// Trigger cadence workerflow
		workflowOptions := cadenceclient.StartWorkflowOptions{
			ID:                              *release.ReleaseID,
			TaskList:                        cadence.ApplicationName,
			ExecutionStartToCloseTimeout:    time.Minute,
			DecisionTaskStartToCloseTimeout: time.Minute,
		}
		h.StartWorkflow(workflowOptions, cadence.Workflow, release)
	default:
		printUsage() // Print how to use and exit
	}
}

func printUsage() {
	fmt.Println("Usage: bifrost json|deploy <release_file> (No args starts Lambda)")
	os.Exit(0)
}
