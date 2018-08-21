package aws

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/sfn"
	"github.com/aws/aws-sdk-go/service/sfn/sfniface"
	ar "github.com/coinbase/step/aws"
)

// S3API aws API
type S3API s3iface.S3API
type SFNAPI sfniface.SFNAPI

type EC2API ec2iface.EC2API

type Clients interface {
	S3(region *string, accountID *string, role *string) S3API
	SFN(region *string, accountID *string, role *string) SFNAPI

	EC2(region *string, accountID *string, role *string) EC2API
}

type ClientsStr struct {
	ar.Clients
}

// S3 returns client for region account and role
func (awsc *ClientsStr) S3(region *string, accountID *string, role *string) S3API {
	return s3.New(awsc.Session(), awsc.Config(region, accountID, role))
}

// SFN is used by the client to init executions
func (awsc *ClientsStr) SFN(region *string, account_id *string, role *string) SFNAPI {
	return sfn.New(awsc.Session(), awsc.Config(region, account_id, role))
}

// S3 returns client for region account and role
func (awsc *ClientsStr) EC2(region *string, accountID *string, role *string) EC2API {
	return ec2.New(awsc.Session(), awsc.Config(region, accountID, role))
}
