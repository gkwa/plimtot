package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
)

// getAWSMetadataHostname returns name of the AWS host from metadata service
func getAWSMetadataAsJson() (string, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal("unable to load config: %w", err)
	}

	client := imds.NewFromConfig(cfg)

	// https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/instancedata-data-retrieval.html

	// ip
	ipRes, err := client.GetMetadata(context.TODO(), &imds.GetMetadataInput{
		Path: "public-ipv4",
	})
	if err != nil {
		log.Fatal("unable to retrieve the ip from the EC2 instance: %w", err)
	}

	defer ipRes.Content.Close()
	ip, err := io.ReadAll(ipRes.Content)
	if err != nil {
		log.Fatal("cannot read ip from the EC2 instance: %w", err)
	}
	log.Printf("ip: %v\n", string(ip))

	// instance-id
	instanceIdRes, err := client.GetMetadata(context.TODO(), &imds.GetMetadataInput{
		Path: "instance-id",
	})
	if err != nil {
		log.Fatal("unable to retrieve the instanceId from the EC2 instance: %w", err)
	}

	defer instanceIdRes.Content.Close()
	instanceId, err := io.ReadAll(instanceIdRes.Content)
	if err != nil {
		log.Fatal("cannot read instanceId from the EC2 instance: %w", err)
	}
	log.Printf("id: %v\n", string(instanceId))

	// region
	region, err := client.GetRegion(context.TODO(), &imds.GetRegionInput{})
	if err != nil {
		log.Printf("Unable to retrieve the region from the EC2 instance %v\n", err)
	}
	log.Printf("region: %v\n", region.Region)

	type Stuff struct {
		Ip         string
		Region     string
		InstanceId string
	}

	myStuff := Stuff{
		Ip:         string(ip),
		Region:     region.Region,
		InstanceId: string(instanceId),
	}

	b, err := json.Marshal(myStuff)
	if err != nil {
		log.Println("error:", err)
	}

	return string(b), nil
}

func main() {
	js, err := getAWSMetadataAsJson()
	if err != nil {
		log.Fatal("cannot read instanceId from the EC2 instance: %w", err)
	}
	os.Stdout.Write([]byte(js))
}
