package deployer

import (
	"bytes"
	"encoding/json"
)

// The goal here is to raise an error if a key is sent that is not supported.
// This should stop many dangerous problems, like misspelling a parameter.

// XRelease is a Release type that can be parsed without overrideing the UnmarshalJSON method
type XRelease Release

// XReleaseExceptions contains the parameters that should not error BUT are not in the release
type XReleaseExceptions struct {
	XRelease
	Task *string // This param is used by step to select the Handler so must be excepted
}

// UnmarshalJSON should error if there is something unexpected
func (release *Release) UnmarshalJSON(data []byte) error {
	var releaseE XReleaseExceptions
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields() // Force error if unknown field is found

	if err := dec.Decode(&releaseE); err != nil {
		return err
	}

	*release = Release(releaseE.XRelease)
	return nil
}
