package mocks

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/coinbase/bifrost/aws"
	"github.com/coinbase/step/utils/to"
)

type DescribeInstancesResponse struct {
	Resp  *ec2.DescribeInstancesOutput
	Error error
}

type TerminateInstancesResponse struct {
	Resp  *ec2.TerminateInstancesOutput
	Error error
}

type DescribeImagesResponse struct {
	Resp  *ec2.DescribeImagesOutput
	Error error
}

type RunInstancesResponse struct {
	Error error
	Resp  *ec2.Reservation
}

// EC2Client returns
type EC2Client struct {
	aws.EC2API
	DescribeInstancesResp  *DescribeInstancesResponse
	TerminateInstancesResp *TerminateInstancesResponse
	DescribeImagesResp     *DescribeImagesResponse
	RunInstancesResp       *RunInstancesResponse
}

func (m *EC2Client) init() {
	if m.DescribeInstancesResp == nil {
		m.DescribeInstancesResp = &DescribeInstancesResponse{
			Resp: &ec2.DescribeInstancesOutput{
				Reservations: []*ec2.Reservation{
					&ec2.Reservation{
						Instances: []*ec2.Instance{},
					},
				},
			},
		}
	}

	if m.TerminateInstancesResp == nil {
		m.TerminateInstancesResp = &TerminateInstancesResponse{}
	}

	if m.DescribeImagesResp == nil {
		m.DescribeImagesResp = &DescribeImagesResponse{}
	}

	if m.RunInstancesResp == nil {
		m.RunInstancesResp = &RunInstancesResponse{
			Resp: &ec2.Reservation{
				Instances: []*ec2.Instance{},
			},
		}
	}
}

func (m *EC2Client) AddImage(image *string) {
	m.DescribeImagesResp = &DescribeImagesResponse{
		Resp: &ec2.DescribeImagesOutput{
			Images: []*ec2.Image{&ec2.Image{ImageId: image}},
		},
	}
}

func createInstance(instanceID *string, project *string, config *string, rid *string, state *string) *ec2.Instance {
	instance := &ec2.Instance{
		Tags: []*ec2.Tag{
			&ec2.Tag{Key: to.Strp("ProjectName"), Value: project},
			&ec2.Tag{Key: to.Strp("ConfgiName"), Value: config},
			&ec2.Tag{Key: to.Strp("ReleaseID"), Value: rid},
		},
	}

	instance.InstanceId = instanceID
	instance.State = &ec2.InstanceState{Name: to.Strp("running")}

	return instance
}

func (m *EC2Client) RunInstance(instanceID *string, project *string, config *string, rid *string) {
	m.init()
	instance := createInstance(instanceID, project, config, rid, to.Strp("ruunning"))
	m.RunInstancesResp.Resp = &ec2.Reservation{
		Instances: []*ec2.Instance{instance},
	}
}

func (m *EC2Client) AddInstance(instanceID *string, project *string, config *string, rid *string, state *string) {
	m.init()

	m.DescribeInstancesResp.Resp.Reservations[0].Instances = append(
		m.DescribeInstancesResp.Resp.Reservations[0].Instances,
		createInstance(instanceID, project, config, rid, state),
	)
}

func (m *EC2Client) DescribeInstances(input *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
	m.init()
	return m.DescribeInstancesResp.Resp, m.DescribeInstancesResp.Error
}

func (m *EC2Client) TerminateInstances(input *ec2.TerminateInstancesInput) (*ec2.TerminateInstancesOutput, error) {
	m.init()
	return m.TerminateInstancesResp.Resp, m.TerminateInstancesResp.Error
}

func (m *EC2Client) RunInstances(input *ec2.RunInstancesInput) (*ec2.Reservation, error) {
	m.init()
	return m.RunInstancesResp.Resp, m.RunInstancesResp.Error
}

func (m *EC2Client) DescribeImages(input *ec2.DescribeImagesInput) (*ec2.DescribeImagesOutput, error) {
	m.init()
	return m.DescribeImagesResp.Resp, m.DescribeImagesResp.Error
}
