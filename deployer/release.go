package deployer

import (
	"fmt"

	"github.com/coinbase/bifrost/aws"
	"github.com/coinbase/bifrost/aws/ec2"
	"github.com/coinbase/step/aws/s3"
	"github.com/coinbase/step/bifrost"
	"github.com/coinbase/step/utils/is"
	"github.com/coinbase/step/utils/to"
)

type Release struct {
	bifrost.Release

	InstanceType *string `json:"instance_type,omitempty"`
	Image        *string `json:"image,omitempty"`

	PreviousInstanceID *string `json:"previous_instance_id,omitempty"`
	NewInstanceID      *string `json:"new_instance_id,omitempty"`

	UserDataSHA256 *string `json:"user_data_sha256,omitempty"`
	userdata       *string // Can be sensitive so stored in S3

	Healthy bool `json:"healthy"`
}

//////////
// Validate
//////////

// Validate returns
func (release *Release) Validate(s3c aws.S3API) error {
	if err := release.Release.Validate(s3c, &Release{}); err != nil {
		return err
	}

	if is.EmptyStr(release.Image) {
		return fmt.Errorf("Image must be defined")
	}

	if is.EmptyStr(release.InstanceType) {
		return fmt.Errorf("InstanceType must be defined")
	}

	if is.EmptyStr(release.UserDataSHA256) {
		return fmt.Errorf("UserDataSHA256 must be defined")
	}

	return nil
}

// ValidateUserDataSHA validates the userdata has the correct SHA for the release
func (release *Release) ValidateUserDataSHA(s3c aws.S3API) error {
	err := release.DownloadUserData(s3c)

	if err != nil {
		return fmt.Errorf("Error Getting UserData with %v", err.Error())
	}

	userdataSha := to.SHA256Str(release.UserData())
	if userdataSha != *release.UserDataSHA256 {
		return fmt.Errorf("UserData SHA incorrect expected %v, got %v", userdataSha, *release.UserDataSHA256)
	}

	return nil
}

// Resource Validations
func (release *Release) ValidateResources(ec2c aws.EC2API) error {
	// There is only one instance to be replaces
	instance, err := ec2.GetInstanceProjectConfig(ec2c, release.ProjectName, release.ConfigName)
	if err != nil {
		return err
	}

	if instance != nil {
		// Set the previous instance to delete
		release.PreviousInstanceID = instance.InstanceId
	}

	// Image Exists
	image, err := ec2.FindImage(ec2c, release.Image)
	if err != nil {
		return err
	}

	if image == nil {
		return fmt.Errorf("Cannot find image %v", *release.Image)
	}

	return nil
}

//////////
// Defaults
//////////

// SetDefaultsWithUserData sets the default values including userdata fetched from S3
func (release *Release) SetDefaults(s3c aws.S3API, region *string, account *string) {

	if release.Timeout == nil {
		release.Timeout = to.Intp(300) // Default to 5 mins
	}

	release.Release.SetDefaults(region, account, "coinbase-bifrost-")

	if is.EmptyStr(release.InstanceType) {
		release.InstanceType = to.Strp("t2.small")
	}
}

//////////
// User Data
//////////

// DownloadUserData fetches and populates the User data from S3
func (release *Release) DownloadUserData(s3c aws.S3API) error {
	if release.userdata != nil {
		return nil
	}

	userdataBytes, err := s3.Get(s3c, release.Bucket, release.UserDataPath())

	if err != nil {
		return err
	}

	release.SetUserData(to.Strp(string(*userdataBytes)))
	return nil
}

// UserDataPath returns
func (release *Release) UserDataPath() *string {
	s := fmt.Sprintf("%v/userdata", *release.ReleaseDir())
	return &s
}

// UserData returns user data
func (release *Release) UserData() *string {
	return release.userdata
}

// SetUserData sets the User data
func (release *Release) SetUserData(userdata *string) {
	release.userdata = userdata
}

//////////
// Deploy/Destroy/Fail
//////////

// SetUserData sets the User data
func (release *Release) Deploy(ec2c aws.EC2API, s3c aws.S3API) error {
	instance := ec2.Instance{}
	instance.AddTag("Name", to.Strp(fmt.Sprintf("%v::%v", *release.ProjectName, *release.ConfigName)))
	instance.AddTag("ProjectName", release.ProjectName)
	instance.AddTag("ConfigName", release.ConfigName)
	instance.AddTag("ReleaseID", release.ReleaseID)

	instance.ImageId = release.Image
	instance.InstanceType = release.InstanceType

	release.DownloadUserData(s3c)

	instanceId, err := instance.Create(ec2c, to.Base64p(release.userdata))
	if err != nil {
		return err
	}

	release.NewInstanceID = instanceId
	return nil
}

func (release *Release) UpdateHealthy(ec2c aws.EC2API) error {
	instance, err := ec2.GetInstanceByID(ec2c, release.NewInstanceID)
	if instance == nil || err != nil {
		return err
	}

	release.Healthy = instance.Healthy()
	return nil
}

func (release *Release) DestroyOldInstance(ec2c aws.EC2API) error {
	if release.PreviousInstanceID == nil {
		return nil
	}
	return ec2.Destroy(ec2c, release.PreviousInstanceID)
}

func (release *Release) DestroyNewInstance(ec2c aws.EC2API) error {
	if release.NewInstanceID == nil {
		return nil
	}
	return ec2.Destroy(ec2c, release.NewInstanceID)
}
