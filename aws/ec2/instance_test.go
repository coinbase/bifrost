package ec2

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/coinbase/bifrost/aws/mocks"
	"github.com/coinbase/step/utils/to"
	"github.com/stretchr/testify/assert"
)

func Test_ToInput(t *testing.T) {
	input := Instance{}.ToInput(to.Strp("asd"))
	assert.NoError(t, input.Validate())
}

func Test_AddTag(t *testing.T) {
	input := Instance{}
	input.AddTag("ProjectName", to.Strp("project"))
	assert.Equal(t, *input.Tags[0].Value, "project")
	assert.Equal(t, *input.Tags[0].Key, "ProjectName")
}

func Test_GetInstanceProjectConfig_NoReservations(t *testing.T) {
	ec2c := mocks.EC2Client{}

	ec2c.DescribeInstancesResp = &mocks.DescribeInstancesResponse{
		Resp: &ec2.DescribeInstancesOutput{
			Reservations: []*ec2.Reservation{},
		},
	}

	output, err := GetInstanceProjectConfig(&ec2c, to.Strp("project"), to.Strp("config"))

	assert.NoError(t, err)
	assert.Nil(t, output)
}

func Test_GetInstanceProjectConfig_NoInstances(t *testing.T) {
	ec2c := mocks.EC2Client{}

	output, err := GetInstanceProjectConfig(&ec2c, to.Strp("project"), to.Strp("config"))

	assert.NoError(t, err)
	assert.Nil(t, output)
}

func Test_GetInstanceProjectConfig_OnceInstance(t *testing.T) {
	ec2c := mocks.EC2Client{}

	ec2c.AddInstance(
		to.Strp("id"),
		to.Strp("project"),
		to.Strp("config"),
		to.Strp("id"),
		to.Strp("running"),
	)

	output, err := GetInstanceProjectConfig(&ec2c, to.Strp("project"), to.Strp("config"))

	assert.NoError(t, err)
	assert.Equal(t, *output.InstanceId, "id")
}

func Test_GetInstanceProjectConfig_TooManyInstance(t *testing.T) {
	ec2c := mocks.EC2Client{}
	ec2c.AddInstance(
		to.Strp("id"),
		to.Strp("project"),
		to.Strp("config"),
		to.Strp("rid"),
		to.Strp("running"),
	)

	ec2c.AddInstance(
		to.Strp("id2"),
		to.Strp("project"),
		to.Strp("config"),
		to.Strp("rid"),
		to.Strp("running"),
	)

	_, err := GetInstanceProjectConfig(&ec2c, to.Strp("project"), to.Strp("config"))

	assert.Error(t, err)
}
