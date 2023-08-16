package iac

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func GetStatus() string {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-west-1"),
	}))

	svc := ec2.New(sess)

	// Describe instances based on the "Name" tag
	input := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("tag:Name"),
				Values: []*string{aws.String("pixelmon")},
			},
		},
	}

	result, err := svc.DescribeInstances(input)
	if err != nil {
		return "Error checking the status."
	}

	if len(result.Reservations) > 0 && len(result.Reservations[0].Instances) > 0 {
		instance := result.Reservations[0].Instances[0]
		if *instance.State.Name == "running" {
			return "The pixelmon EC2 instance is online."
		} else {
			return "The pixelmon EC2 instance is not online."
		}
	}

	return "The pixelmon EC2 instance was not found."
}

func Start() string {
	return "Starting..."
}

func Stop() string {
	return "Stopping..."
}
