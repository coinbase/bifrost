package deployer

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/coinbase/bifrost/aws/mocks"
	"github.com/coinbase/step/bifrost"
	"github.com/coinbase/step/machine"
	"github.com/coinbase/step/utils/to"
	"github.com/stretchr/testify/assert"
)

////////
// RELEASE
////////

func MockRelease() *Release {
	release := &Release{
		Release: bifrost.Release{
			AwsAccountID: to.Strp("00000000"),
			ReleaseID:    to.Strp("release-1"),
			ProjectName:  to.Strp("project"),
			ConfigName:   to.Strp("development"),
			CreatedAt:    to.Timep(time.Now()),
		},
		InstanceType:   to.Strp("t2.small"),
		Image:          to.Strp("ami-123"),
		UserDataSHA256: to.Strp("sha256"),
	}

	if release.UserData() == nil {
		release.SetUserData(to.Strp("#cloud_config"))
	}
	release.UserDataSHA256 = to.Strp(to.SHA256Str(release.UserData()))

	return release
}

func MockAwsClients(release *Release) *mocks.MockClients {
	awsc := mocks.MockAWS()

	raw, _ := json.Marshal(release)
	accountID := release.AwsAccountID
	if accountID == nil {
		accountID = to.Strp("000000000000")
	}

	if release.ProjectName != nil && release.ConfigName != nil && release.ReleaseID != nil {
		releasePath := fmt.Sprintf("%v/%v/%v/%v/release", *accountID, *release.ProjectName, *release.ConfigName, *release.ReleaseID)
		awsc.S3Client.AddGetObject(releasePath, string(raw), nil)

		if release.UserData() != nil {
			awsc.S3Client.AddGetObject(*release.UserDataPath(), *release.UserData(), nil)
		}

		awsc.EC2Client.AddImage(release.Image)
		// return this as the instance ID
		awsc.EC2Client.RunInstance(to.Strp("instanceid"), release.ProjectName, release.ConfigName, release.ReleaseID)
		// return the healthy
		awsc.EC2Client.AddInstance(to.Strp("instanceid"), release.ProjectName, release.ConfigName, release.ReleaseID, to.Strp("running"))
	}

	return awsc
}

func createTestStateMachine(t *testing.T, awsc *mocks.MockClients) *machine.StateMachine {
	stateMachine, err := StateMachineWithTaskHandlers(CreateTaskFunctinons(awsc))
	assert.NoError(t, err)

	return stateMachine
}

func assertSuccessfulExecution(t *testing.T, release *Release) {
	stateMachine := createTestStateMachine(t, MockAwsClients(release))

	exec, err := stateMachine.Execute(release)
	output := exec.Output

	assert.NoError(t, err)
	assert.Equal(t, true, output["success"])
	assert.NotRegexp(t, "error", exec.LastOutputJSON)

	assert.Equal(t, exec.Path(), []string{
		"Validate",
		"Lock",
		"ValidateResources",
		"Deploy",
		"WaitForDeploy",
		"WaitForHealthy",
		"CheckHealthy",
		"Healthy?",
		"CleanUpSuccess",
		"Success",
	})
}
