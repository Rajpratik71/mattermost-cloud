package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/stretchr/testify/suite"

	"github.com/mattermost/mattermost-cloud/internal/tools/aws/mocks"
)

// ClientTestSuite supplies tests for aws package Client.
type ClientTestSuite struct {
	suite.Suite
	session *session.Session
}

func (d *ClientTestSuite) SetupTest() {
	d.session = session.Must(session.NewSession())
}

func (d *ClientTestSuite) TestNewClient() {
	client := NewClient(d.session)

	d.Assert().NotNil(client)
	d.Assert().NotNil(client.RDS)
	d.Assert().NotNil(client.S3)
	d.Assert().NotNil(client.IAM)
	d.Assert().NotNil(client.Route53)
	d.Assert().NotNil(client.ACM)
}

func TestClientSuite(t *testing.T) {
	suite.Run(t, new(DatabaseMigrationTestSuite))
}

// MockedClient supplies a AWS mocks and mocked AWS client.
type MockedClient struct {
	api    *mocks.Mocks
	client *Client
}

// NewMockedClient returns a instance of a mocked AWS client.
func NewMockedClient() *MockedClient {
	mockedClient := &MockedClient{
		api: mocks.NewMocks(),
	}
	mockedClient.client = &Client{
		RDS:     mockedClient.api.RDS,
		EC2:     mockedClient.api.EC2,
		IAM:     mockedClient.api.IAM,
		ACM:     mockedClient.api.ACM,
		S3:      mockedClient.api.S3,
		Route53: mockedClient.api.Route53,
	}

	return mockedClient
}