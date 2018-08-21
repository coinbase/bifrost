package ec2

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/coinbase/bifrost/aws"
)

// Image struct
type Image struct {
	ImageID *string
}

// FindImage takes either a ID or a Tag of an ami e.g. ubuntu or ami-00000000
func FindImage(ec2c aws.EC2API, id *string) (*Image, error) {
	output, err := ec2c.DescribeImages(&ec2.DescribeImagesInput{
		ImageIds: []*string{id},
	})

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	switch len(output.Images) {
	case 0:
		return nil, nil
	case 1:
		im := output.Images[0]
		if im == nil {
			return nil, fmt.Errorf("AMI Image nil")
		}
		return &Image{im.ImageId}, nil
	default:
		return nil, fmt.Errorf("Should only be one error")
	}
}
