package deployer

import (
	"testing"

	"github.com/coinbase/step/utils/to"
	fuzz "github.com/google/gofuzz"
	"github.com/stretchr/testify/assert"
)

// Test_Release_Basic_Fuzz
func Test_Release_Basic_Fuzz(t *testing.T) {
	for i := 0; i < 50; i++ {
		f := fuzz.New()
		var release Release
		f.Fuzz(&release)

		assertNoPanic(t, &release)
	}
}

func assertNoPanic(t *testing.T, release *Release) {
	release.AwsAccountID = to.Strp("0000000")
	stateMachine := createTestStateMachine(t, MockAwsClients(release))

	exec, err := stateMachine.Execute(release)
	if err != nil {
		assert.NotRegexp(t, "Panic", err.Error())
	}

	assert.NotNil(t, exec)
	assert.NotRegexp(t, "Panic", exec.LastOutputJSON)
}
