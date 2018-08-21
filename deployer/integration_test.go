package deployer

import (
	"testing"

	"github.com/coinbase/step/utils/to"
	"github.com/stretchr/testify/assert"
)

///////////////
// Successful Tests
///////////////

func Test_Successful_Execution_Works(t *testing.T) {
	release := MockRelease()
	assertSuccessfulExecution(t, release)
}

///////////////
// Unsuccessful Tests
///////////////

func Test_Unsuccessful_Bad_Userdata_SHA(t *testing.T) {
	release := MockRelease()
	stateMachine := createTestStateMachine(t, MockAwsClients(release))

	release.UserDataSHA256 = to.Strp("bad sha")

	exec, err := stateMachine.Execute(release)

	assert.Error(t, err)
	assert.Regexp(t, "BadReleaseError", exec.LastOutputJSON)

	assert.Equal(t, []string{
		"Validate",
		"FailureClean",
	}, exec.Path())
}

func Test_Unsuccessful_Bad_Input(t *testing.T) {
	stateMachine := createTestStateMachine(t, MockAwsClients(MockRelease()))
	exec, err := stateMachine.Execute(`{"input": "bad"}`)

	assert.Error(t, err)
	assert.Regexp(t, "UnmarshalError", exec.LastOutputJSON)

	assert.Equal(t, []string{
		"Validate",
		"FailureClean",
	}, exec.Path())
}
