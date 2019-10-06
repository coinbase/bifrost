package client

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/sfn/sfniface"
	"github.com/coinbase/bifrost/aws"
	"github.com/coinbase/bifrost/deployer"
	"github.com/coinbase/step/aws/s3"
	"github.com/coinbase/step/execution"
	"github.com/coinbase/step/utils/to"
)

// Deploy attempts to deploy release
func Deploy(step_fn *string, releaseFile *string) error {
	region, accountID := to.RegionAccount()
	release, err := ReleaseFromFile(releaseFile, region, accountID)
	if err != nil {
		return err
	}

	deployerARN := to.StepArn(region, accountID, step_fn)

	return deploy(&aws.ClientsStr{}, release, deployerARN)
}

func DeployS3(awsc aws.Clients, releaseFile *string) (*deployer.Release, error) {
	region, accountID := to.RegionAccount()
	release, err := ReleaseFromFile(releaseFile, region, accountID)
	if err != nil {
		return nil, err
	}

	// Uploading the Release to S3 to match SHAs
	if err := s3.PutStruct(awsc.S3(nil, nil, nil), release.Bucket, release.ReleasePath(), release); err != nil {
		return nil, err
	}

	// Uploading the encrypted Userdata to S3
	if err := s3.PutSecure(awsc.S3(nil, nil, nil), release.Bucket, release.UserDataPath(), release.UserData(), kMSKey()); err != nil {
		return nil, err
	}

	return release, nil
}

func kMSKey() *string {
	// TODO: allow customization of the KMS key from the command line utility
	return to.Strp("alias/aws/s3")
}

func deploy(awsc aws.Clients, release *deployer.Release, deployerARN *string) error {
	// Uploading the Release to S3 to match SHAs
	if err := s3.PutStruct(awsc.S3(nil, nil, nil), release.Bucket, release.ReleasePath(), release); err != nil {
		return err
	}

	// Uploading the encrypted Userdata to S3
	if err := s3.PutSecure(awsc.S3(nil, nil, nil), release.Bucket, release.UserDataPath(), release.UserData(), kMSKey()); err != nil {
		return err
	}

	exec, err := findOrCreateExec(awsc.SFN(nil, nil, nil), deployerARN, release)
	if err != nil {
		return err
	}

	// Execute every second
	exec.WaitForExecution(awsc.SFN(nil, nil, nil), 1, waiter)
	fmt.Println("")
	return nil
}

func findOrCreateExec(sfnc sfniface.SFNAPI, deployer *string, release *deployer.Release) (*execution.Execution, error) {
	exec, err := execution.FindExecution(sfnc, deployer, release.ExecutionPrefix())
	if err != nil {
		return nil, err
	}

	if exec != nil {
		return exec, nil
	}

	return execution.StartExecution(sfnc, deployer, release.ExecutionName(), release)
}
