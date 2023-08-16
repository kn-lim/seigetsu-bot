package pixelmon

import (
	"errors"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func GetStatus() string {
	svc, err := getEC2Session()
	if err != nil {
		return err.Error()
	}

	result, err := getPixelmonServer(svc)
	if err != nil {
		return Message[Err_Status]
	}

	if len(result.Reservations) > 0 && len(result.Reservations[0].Instances) > 0 {
		instance := result.Reservations[0].Instances[0]
		if *instance.State.Name == "running" {
			return Message[Online]
		} else {
			return Message[Offline]
		}
	}

	return Message[Not_Found]
}

func Start() string {
	svc, err := getEC2Session()
	if err != nil {
		return err.Error()
	}

	result, err := getPixelmonServer(svc)
	if err != nil {
		return Message[Err_Status]
	}

	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			if *instance.State.Name == "running" {
				return Message[Online]
			} else if *instance.State.Name == "stopped" {
				startInput := &ec2.StartInstancesInput{
					InstanceIds: []*string{instance.InstanceId},
				}

				_, err := svc.StartInstances(startInput)
				if err != nil {
					log.Printf("Failed to start pixelmon: %v", err)
					return Message[Err_Start]
				}

				return Message[Starting]
			}
		}
	}

	return Message[Not_Found]
}

func Stop() string {
	svc, err := getEC2Session()
	if err != nil {
		return err.Error()
	}

	result, err := getPixelmonServer(svc)
	if err != nil {
		return Message[Err_Status]
	}

	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			if *instance.State.Name == "stopped" {
				return Message[Offline]
			} else if *instance.State.Name == "running" {
				stopInput := &ec2.StopInstancesInput{
					InstanceIds: []*string{instance.InstanceId},
				}

				_, err := svc.StopInstances(stopInput)
				if err != nil {
					log.Printf("Failed to stop pixelmon: %v", err)
					return Message[Err_Stop]
				}

				return Message[Stopping]
			}
		}
	}

	return Message[Not_Found]
}

func getEC2Session() (*ec2.EC2, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-west-1"),
	})
	if err != nil {
		log.Printf("Error creating AWS session: %v", err)
		return nil, errors.New("Error creating AWS session")
	}

	return ec2.New(sess), nil
}

func getPixelmonServer(svc *ec2.EC2) (*ec2.DescribeInstancesOutput, error) {
	// Describe instances based on the "Name" tag
	input := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("tag:Name"),
				Values: []*string{aws.String("pixelmon")},
			},
		},
	}

	return svc.DescribeInstances(input)
}
