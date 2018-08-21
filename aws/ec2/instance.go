package ec2

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/coinbase/bifrost/aws"
	"github.com/coinbase/step/utils/to"
)

type Instance struct {
	ec2.Instance
}

// GetInstnaces Returns Instance for project config
// errors if more than one instance is returned
func GetInstanceProjectConfig(ec2c aws.EC2API, project *string, config *string) (*Instance, error) {
	filters := []*ec2.Filter{
		&ec2.Filter{
			Name:   to.Strp("tag:ProjectName"),
			Values: []*string{project},
		},
		&ec2.Filter{
			Name:   to.Strp("tag:ConfigName"),
			Values: []*string{config},
		},
	}
	return GetInstance(ec2c, &ec2.DescribeInstancesInput{Filters: filters})
}

func GetInstanceReleaseID(ec2c aws.EC2API, releaseID *string) (*Instance, error) {
	filters := []*ec2.Filter{
		&ec2.Filter{
			Name:   to.Strp("tag:ReleaseID"),
			Values: []*string{releaseID},
		},
	}

	return GetInstance(ec2c, &ec2.DescribeInstancesInput{Filters: filters})
}

func GetInstanceByID(ec2c aws.EC2API, instanceID *string) (*Instance, error) {
	return GetInstance(ec2c, &ec2.DescribeInstancesInput{InstanceIds: []*string{instanceID}})
}

func GetInstance(ec2c aws.EC2API, input *ec2.DescribeInstancesInput) (*Instance, error) {
	output, err := ec2c.DescribeInstances(input)

	if err != nil {
		return nil, err
	}

	allInstances := []*ec2.Instance{}

	if output == nil || output.Reservations == nil {
		return nil, nil
	}

	for _, r := range output.Reservations {
		allInstances = append(allInstances, r.Instances...)
	}

	runningInstances := []*Instance{}
	for _, awsInstance := range allInstances {
		i := &Instance{*awsInstance}
		if i.Exists() {
			runningInstances = append(runningInstances, i)
		}
	}

	switch len(runningInstances) {
	case 0:
		return nil, nil
	case 1:
		return runningInstances[0], nil
	default:
		return nil, fmt.Errorf("Found to many instances")
	}
}

func Destroy(ec2c aws.EC2API, instanceId *string) error {
	_, err := ec2c.TerminateInstances(&ec2.TerminateInstancesInput{
		InstanceIds: []*string{instanceId},
	})
	return err
}

////
// Instance Methods
////

// State methods
// STATES: pending | running | shutting-down | terminated | stopping | stopped
func (i *Instance) Healthy() bool {
	if i.State == nil || i.State.Name == nil {
		return false
	}
	return *i.State.Name == "running"
}

func (i *Instance) Stopping() bool {
	if i.State == nil || i.State.Name == nil {
		return false
	}

	return *i.State.Name != "pending" && *i.State.Name != "running"
}

// Exists returns if the instance will exist soon.
func (i *Instance) Exists() bool {
	if i.State == nil || i.State.Name == nil {
		return false
	}

	return *i.State.Name != "shutting-down" && *i.State.Name != "terminated"
}

// AddTag adds a tag to the input
func (s *Instance) AddTag(key string, value *string) {
	if s.Tags == nil {
		s.Tags = []*ec2.Tag{}
	}

	for _, tag := range s.Tags {
		if *tag.Key == key {
			tag.Value = value
			return // Found the tag key already
		}
	}

	// Add new Tag
	s.Tags = append(s.Tags, &ec2.Tag{Key: &key, Value: value})
}

func (i Instance) ToInput(userData *string) *ec2.RunInstancesInput {
	return &ec2.RunInstancesInput{
		SubnetId:     i.SubnetId,
		ImageId:      i.ImageId,
		InstanceType: i.InstanceType,
		UserData:     userData,
		MinCount:     to.Int64p(1),
		MaxCount:     to.Int64p(1),
		TagSpecifications: []*ec2.TagSpecification{&ec2.TagSpecification{
			Tags:         i.Tags,
			ResourceType: to.Strp("instance"),
		}},
	}
}

func (i Instance) Create(ec2c aws.EC2API, userData *string) (*string, error) {
	reservation, err := ec2c.RunInstances(i.ToInput(userData))
	if err != nil {
		return nil, err
	}

	if len(reservation.Instances) != 1 {
		return nil, fmt.Errorf("Unknown Error")
	}

	i.Instance = *reservation.Instances[0]

	return i.Instance.InstanceId, nil
}
