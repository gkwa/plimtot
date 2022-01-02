package slow

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
)

func getCredentialsFromRole() (*credentials.Credentials, error) {
	roleProvider := &ec2rolecreds.EC2RoleProvider{
		Client: ec2metadata.New(session.New()),
	}
	creds := credentials.NewCredentials(roleProvider)

	start := time.Now().UTC()
	if _, err := creds.Get(); err != nil { // this takes 20 seconds
		return nil, err
	}
	fmt.Printf("getting credentails from role took %s\n", time.Now().UTC().Sub(start))

	return creds, nil
}
