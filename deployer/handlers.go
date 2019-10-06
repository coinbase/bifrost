package deployer

import (
	"context"

	"github.com/coinbase/bifrost/aws"
	"github.com/coinbase/step/errors"
	"github.com/coinbase/step/utils/to"
)

// DeployHandler function type
type DeployHandler func(context.Context, *Release) (*Release, error)

////////////
// HANDLERS
////////////

var assumedRole = to.Strp("coinbase-bifrost-assumed")

// Validate checks the release
func Validate(awsc aws.Clients) DeployHandler {
	return func(ctx context.Context, release *Release) (*Release, error) {
		// Assign the release its SHA before anything alters it
		release.ReleaseSHA256 = to.SHA256Struct(release)

		// Default the releases Account and Region to where the Lambda is running
		region, account := to.AwsRegionAccountFromContext(ctx)

		// Fill in all the blank Attributes
		release.SetDefaults(awsc.S3(nil, nil, nil), region, account)

		if err := release.Validate(awsc.S3(nil, nil, nil)); err != nil {
			return nil, &errors.BadReleaseError{err.Error()}
		}

		return release, nil
	}
}

// Lock secures a lock for the release
func Lock(awsc aws.Clients) DeployHandler {
	return func(ctx context.Context, release *Release) (*Release, error) {
		// returns LockExistsError, LockError
		return release, release.GrabLock(awsc.S3(nil, nil, nil))
	}
}

// ValidateResources calls to AWS to make sure all references resources exist
func ValidateResources(awsc aws.Clients) DeployHandler {
	return func(ctx context.Context, release *Release) (*Release, error) {

		if err := release.ValidateResources(awsc.EC2(release.AwsRegion, release.AwsAccountID, assumedRole)); err != nil {
			return nil, errors.BadReleaseError{err.Error()}
		}

		return release, nil
	}
}

// Deploy crates new AWS resources for the release
func Deploy(awsc aws.Clients) DeployHandler {
	return func(_ context.Context, release *Release) (*Release, error) {

		if err := release.Deploy(
			awsc.EC2(release.AwsRegion, release.AwsAccountID, assumedRole),
			awsc.S3(nil, nil, nil),
		); err != nil {
			return nil, &errors.BadReleaseError{err.Error()}
		}

		return release, nil
	}
}

// CheckHealthy checks all new resources are healthy
func CheckHealthy(awsc aws.Clients) DeployHandler {
	return func(_ context.Context, release *Release) (*Release, error) {

		// Checks if Timeout has been reached & if the halt flag has been uploaded
		if err := release.IsHalt(awsc.S3(nil, nil, nil)); err != nil {
			return nil, &errors.HaltError{err.Error()}
		}

		if err := release.UpdateHealthy(awsc.EC2(release.AwsRegion, release.AwsAccountID, assumedRole)); err != nil {
			return nil, &errors.HaltError{err.Error()}
		}

		return release, nil
	}
}

// CleanUpSuccess deletes old deploy resources and releases the lock
func CleanUpSuccess(awsc aws.Clients) DeployHandler {
	return func(_ context.Context, release *Release) (*Release, error) {

		// This will destroy the old instance that is running
		// We are ignoring if this errors because if we error we might end up deleting ALL instances
		// TODO: properly handle if this breaks here
		release.DestroyOldInstance(awsc.EC2(release.AwsRegion, release.AwsAccountID, assumedRole))

		if err := release.ReleaseLock(awsc.S3(nil, nil, nil)); err != nil {
			return nil, &errors.LockError{err.Error()}
		}

		release.Success = to.Boolp(true)

		return release, nil
	}
}

// CleanUpFailure releases the lock then fails
func CleanUpFailure(awsc aws.Clients) DeployHandler {
	return func(_ context.Context, release *Release) (*Release, error) {

		if err := release.DestroyNewInstance(awsc.EC2(release.AwsRegion, release.AwsAccountID, assumedRole)); err != nil {
			return nil, &errors.CleanUpError{err.Error()}
		}

		if err := release.ReleaseLock(awsc.S3(nil, nil, nil)); err != nil {
			return nil, &errors.LockError{err.Error()}
		}

		return release, nil
	}
}
