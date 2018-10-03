package main

import (
	"fmt"
	"os"

	"github.com/coinbase/bifrost/client"
	"github.com/coinbase/bifrost/deployer"
	"github.com/coinbase/step/utils/run"
	"github.com/coinbase/step/utils/to"
)

func main() {
	var arg, command string
	switch len(os.Args) {
	case 1:
		fmt.Println("Starting Lambda")
		run.LambdaTasks(deployer.TaskFunctions())
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

	switch command {
	case "json":
		// This is used to bootstrap the state-machine
		run.JSON(deployer.StateMachine())
	case "deploy":
		// Send Configuration to the deployer
		// arg is a filename
		err := client.Deploy(step_fn, &arg)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	default:
		printUsage() // Print how to use and exit
	}
}

func printUsage() {
	fmt.Println("Usage: bifrost json|deploy <release_file> (No args starts Lambda)")
	os.Exit(0)
}
