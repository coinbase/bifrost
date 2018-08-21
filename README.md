# Bifrost

Bifrost is a set of conventions for building a deployer using the Step framework. The goal of this repository is to be a template, a "paved road", to building a deployer quickly and safely. This is accomplished by implementing a demo-deployer that deploys a single EC2 instance (**NOT FOR PRODUCTION**).

### Folder Structure

Here is a quick breakdown to the files in this repository

```
bifrost/
├── .circleci/
│   └── config.yml # example CI setup
├── aws/
│   ├── ec2/ # example ec2 client
│   ├── mocks/ # example mock clients
│   └── aws.go # setup for aws.Clients
├── client/
│   └── client.go # example client code
├── deployer/
│   ├── fuzz_test.go # Fuzz Test example
│   ├── integration_test.go # Tests for the Deployer
│   ├── machine.go # state-machine definition
│   ├── handlers.go # Handler functions for states
│   └── release.go # Bifrost Release
├── releases/
│   └── release.json # example release
├── scripts/
│   └── bootstrap_deplyer # bootstraping script
├── bifrost.go # executable code
├── Gopkg.toml # Go dependencies
└── Dockerfile # Build bifrost for deploy
```

Each folder has a README.md that contains a description of what should exist in that folder.

### Getting Started

The way to use this is to execute in the correct folder in your GOPATH:

```
export ORG=<your-org>
export DEPLOYER=<your_deployer>

git clone git@github.com:coinbase/bifrost.git $DEPLOYER
cd $DEPLOYER

scripts/rename
```

This can be tested by ensuring tests pass with `go test ./...` and Docker builds `docker build`: