package cadence

import (
	"context"
	"time"
	"fmt"

	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
	"go.uber.org/cadence"
	"go.uber.org/zap"

	"github.com/coinbase/bifrost/deployer"
	"github.com/coinbase/bifrost/aws"
)

/**
 * This is the hello world workflow sample.
 */

// ApplicationName is the task list for this sample
const ApplicationName = "bifrostDeployer"

// This is registration process where you register all your workflows
// and activity function handlers.
func init() {
	workflow.Register(Workflow)
	activity.Register(validate)
	activity.Register(lock)
	activity.Register(validateResources)
	activity.Register(cleanupFailure)
	activity.Register(cleanupSuccess)
	activity.Register(deploy)
	activity.Register(checkHealthy)
}

// Workflow workflow decider
func Workflow(ctx workflow.Context, release *deployer.Release) error {
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		HeartbeatTimeout:       time.Second * 20,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("workflow started")
	err := workflow.ExecuteActivity(ctx, validate, release).Get(ctx, &release)
	if err != nil {
		logger.Error("Validate activity failed.", zap.Error(err))
		return err
	}

	// TODO: can a cadence mutex be used here?
	logger.Info("lock activity starting")
	err = workflow.ExecuteActivity(ctx, lock, release).Get(ctx, &release)
	if err != nil {
		logger.Error("lock activity failed.", zap.Error(err))
		return err
	}

	logger.Info("validateResources activity starting")
	err = workflow.ExecuteActivity(ctx, validateResources, release).Get(ctx, &release)
	if err != nil {
		logger.Error("validateResources activity failed.", zap.Error(err))

		logger.Info("cleanupFailure activity starting")
		err = workflow.ExecuteActivity(ctx, cleanupFailure, release).Get(ctx, &release)
		if err != nil {
			logger.Error("cleanupFailure activity failed.", zap.Error(err))
		}

		return err
	}

	logger.Info("deploy activity starting")
	err = workflow.ExecuteActivity(ctx, deploy, release).Get(ctx, &release)
	if err != nil {
		logger.Error("deploy activity failed.", zap.Error(err))

		logger.Info("cleanupFailure activity starting")
		err = workflow.ExecuteActivity(ctx, cleanupFailure, release).Get(ctx, &release)
		if err != nil {
			logger.Error("cleanupFailure activity failed.", zap.Error(err))
		}

		return err
	}

	logger.Info("checkHealthy activity starting")
	healthyAo := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    5 * time.Minute,
		HeartbeatTimeout:       time.Second * 20,
		RetryPolicy: &cadence.RetryPolicy{
			InitialInterval: 10 * time.Second,
			MaximumInterval: 10 * time.Second,
			ExpirationInterval: 5*time.Minute,
		},
	}
	healthyAoCtx := workflow.WithActivityOptions(ctx, healthyAo)
	err = workflow.ExecuteActivity(healthyAoCtx, checkHealthy, release).Get(ctx, &release)
	if err != nil {
		logger.Error("deploy activity failed.", zap.Error(err))

		logger.Info("cleanupFailure activity starting")
		err = workflow.ExecuteActivity(ctx, cleanupFailure, release).Get(ctx, &release)
		if err != nil {
			logger.Error("cleanupFailure activity failed.", zap.Error(err))
		}

		return err
	}


	logger.Info("cleanupSuccess activity starting")
	err = workflow.ExecuteActivity(ctx, cleanupSuccess, release).Get(ctx, &release)
	if err != nil {
		logger.Error("cleanupSuccess activity failed.", zap.Error(err))
		return err
	}

	logger.Info("Workflow completed.", zap.Any("release", release))

	return nil
}

func validate(ctx context.Context, release *deployer.Release) (*deployer.Release, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("validate activity started", zap.String("release_id", *release.ReleaseID))

	clients := &aws.ClientsStr{}
	release, err := deployer.Validate(clients)(ctx, release)
	if err != nil {
		return nil, err
	}

	return release, nil
}

func lock(ctx context.Context, release *deployer.Release) (*deployer.Release, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("lock activity started", zap.String("release_id", *release.ReleaseID))

	clients := &aws.ClientsStr{}
	release, err := deployer.Lock(clients)(ctx, release)
	if err != nil {
		return nil, err
	}

	return release, nil
}

func validateResources(ctx context.Context, release *deployer.Release) (*deployer.Release, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("validateResources activity started", zap.String("release_id", *release.ReleaseID))

	clients := &aws.ClientsStr{}
	release, err := deployer.ValidateResources(clients)(ctx, release)
	if err != nil {
		return nil, err
	}

	return release, nil
}

func deploy(ctx context.Context, release *deployer.Release) (*deployer.Release, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("deploy activity started", zap.String("release_id", *release.ReleaseID))

	clients := &aws.ClientsStr{}
	release, err := deployer.Deploy(clients)(ctx, release)
	if err != nil {
		return nil, err
	}

	return release, nil
}

func cleanupFailure(ctx context.Context, release *deployer.Release) (*deployer.Release, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("cleanupFailure activity started", zap.String("release_id", *release.ReleaseID))

	clients := &aws.ClientsStr{}
	release, err := deployer.CleanUpFailure(clients)(ctx, release)
	if err != nil {
		return nil, err
	}

	return release, nil
}

func cleanupSuccess(ctx context.Context, release *deployer.Release) (*deployer.Release, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("cleanupSuccess activity started", zap.String("release_id", *release.ReleaseID))

	clients := &aws.ClientsStr{}
	release, err := deployer.CleanUpSuccess(clients)(ctx, release)
	if err != nil {
		return nil, err
	}

	return release, nil
}

func checkHealthy(ctx context.Context, release *deployer.Release) (*deployer.Release, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("checkHealthy activity started", zap.String("release_id", *release.ReleaseID))

	clients := &aws.ClientsStr{}
	release, err := deployer.CheckHealthy(clients)(ctx, release)
	if err != nil {
		return nil, err
	}

	if !release.Healthy {
		return nil, fmt.Errorf("deploy is not healthy")
	}

	return release, nil
}
