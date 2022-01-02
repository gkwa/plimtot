package main

import (
  "fmt"
  "github.com/aws/aws-sdk-go-v2/aws/external"
  "github.com/aws/aws-sdk-go-v2/aws/ec2metadata"
)

func main() {
  cfg, err := external.LoadDefaultAWSConfig()

  if err != nil {
    panic("Unable to load SDK config, " + err.Error())
  }

  md_svc := ec2metadata.New(cfg)

  if !md_svc.Available() {
    panic("Metadata service cannot be reached.  Are you on an EC2/ECS/Lambda machine?")
  }

  region, err := md_svc.Region()

  fmt.Println("Region is: " + region)
}
