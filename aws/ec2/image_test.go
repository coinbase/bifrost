package ec2

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/coinbase/bifrost/aws/mocks"
	"github.com/coinbase/step/utils/to"
	"github.com/stretchr/testify/assert"
)

func Test_FindImage_ErrorsNoImage(t *testing.T) {
	ec2c := mocks.EC2Client{}

	ec2c.DescribeImagesResp = &mocks.DescribeImagesResponse{
		Resp: &ec2.DescribeImagesOutput{
			Images: []*ec2.Image{},
		},
	}

	output, err := FindImage(&ec2c, to.Strp("asd"))

	assert.NoError(t, err)
	assert.Nil(t, output)
}

func Test_FindImage_Works(t *testing.T) {
	ec2c := mocks.EC2Client{}

	ec2c.DescribeImagesResp = &mocks.DescribeImagesResponse{
		Resp: &ec2.DescribeImagesOutput{
			Images: []*ec2.Image{&ec2.Image{ImageId: to.Strp("qwe")}},
		},
	}

	output, err := FindImage(&ec2c, to.Strp("asd"))

	assert.NoError(t, err)
	assert.Equal(t, *output.ImageID, "qwe")
}
