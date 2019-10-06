package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/coinbase/bifrost/deployer"
	"github.com/coinbase/step/bifrost"
	"github.com/coinbase/step/execution"
	"github.com/coinbase/step/utils/to"
)

func prepareRelease(release *deployer.Release, region *string, accountID *string) {
	release.Release.SetDefaults(region, accountID, "coinbase-bifrost-")

	release.ReleaseID = to.TimeUUID("release-")
	release.CreatedAt = to.Timep(time.Now())
}

func parseRelease(releaseFile string) (*deployer.Release, error) {
	rawRelease, err := ioutil.ReadFile(releaseFile)
	if err != nil {
		return nil, err
	}

	var release deployer.Release
	if err := json.Unmarshal(rawRelease, &release); err != nil {
		return nil, err
	}

	return &release, nil
}

func parseUserData(releaseFile string) (*string, error) {
	userdataFile := fmt.Sprintf("%v.userdata", releaseFile)
	rawUserData, err := ioutil.ReadFile(userdataFile)

	if err != nil {
		return nil, err
	}

	return to.Strp(string(rawUserData)), nil
}

func ReleaseFromFile(releaseFile *string, region *string, accountID *string) (*deployer.Release, error) {
	release, err := parseRelease(*releaseFile)
	if err != nil {
		return nil, err
	}

	userdata, err := parseUserData(*releaseFile)
	if err != nil {
		return nil, err
	}

	release.SetUserData(userdata)
	release.UserDataSHA256 = to.Strp(to.SHA256Str(userdata))

	prepareRelease(release, region, accountID)

	return release, nil
}

func waiter(ed *execution.Execution, sd *execution.StateDetails, err error) error {
	if err != nil {
		return fmt.Errorf("Unexpected Error %v", err.Error())
	}

	var release_error struct {
		Error *bifrost.ReleaseError `json:"error,omitempty"`
	}

	if sd != nil && sd.LastOutput != nil {
		json.Unmarshal([]byte(*sd.LastOutput), &release_error)
	}

	fmt.Printf("\rExecution: %v", *ed.Status)
	if release_error.Error != nil {
		fmt.Printf("\nError: %v\nCause: %v\n", to.Strs(release_error.Error.Error), to.Strs(release_error.Error.Cause))
	}

	return nil
}
