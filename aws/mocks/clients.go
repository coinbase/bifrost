package mocks

import (
	"github.com/coinbase/bifrost/aws"
	"github.com/coinbase/step/aws/mocks"
)

// MockClients struct
type MockClients struct {
	S3Client  *mocks.MockS3Client
	SFNClient *mocks.MockSFNClient

	EC2Client *EC2Client
}

func (m MockClients) S3(region *string, accountID *string, role *string) aws.S3API {
	return m.S3Client
}

func (m MockClients) SFN(*string, *string, *string) aws.SFNAPI {
	return m.SFNClient
}

func (m MockClients) EC2(region *string, accountID *string, role *string) aws.EC2API {
	return m.EC2Client
}

// MockAWS mock clients
func MockAWS() *MockClients {
	return &MockClients{
		S3Client:  &mocks.MockS3Client{},
		SFNClient: &mocks.MockSFNClient{},

		EC2Client: &EC2Client{},
	}
}
