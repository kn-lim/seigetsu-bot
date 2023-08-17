package pixelmon

import (
	"errors"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ssm"
)

func GetStatus() string {
	sess, err := getSession()
	if err != nil {
		return err.Error()
	}

	svc := ec2.New(sess)

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
	sess, err := getSession()
	if err != nil {
		return err.Error()
	}

	svc := ec2.New(sess)

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

				PixelmonInstanceID = *instance.InstanceId

				return StartPixelmonService()
			}
		}
	}

	return Message[Not_Found]
}

func Stop() string {
	sess, err := getSession()
	if err != nil {
		return err.Error()
	}

	svc := ec2.New(sess)

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

func StartPixelmonService() string {
	sess, err := getSession()
	if err != nil {
		return err.Error()
	}

	ec2Svc := ec2.New(sess)

	for {
		input := &ec2.DescribeInstancesInput{
			InstanceIds: []*string{aws.String(PixelmonInstanceID)},
		}

		result, err := ec2Svc.DescribeInstances(input)
		if err != nil {
			return err.Error()
		}

		if len(result.Reservations) == 0 || len(result.Reservations[0].Instances) == 0 {
			return Message[Not_Found]
		}

		state := result.Reservations[0].Instances[0].State.Name
		if *state == ec2.InstanceStateNameRunning {
			break
		}

		time.Sleep(10 * time.Second)
	}

	ssmSvc := ssm.New(sess)

	document := "AWS-RunShellScript"
	commands := []string{"screen -S mc -d -m 'java -Xms3G -Xmx3G -XX:+UseG1GC -XX:+UnlockExperimentalVMOptions -XX:MaxGCPauseMillis=100 -XX:+DisableExplicitGC -XX:TargetSurvivorRatio=90 -XX:G1NewSizePercent=50 -XX:G1MaxNewSizePercent=80 -XX:G1MixedGCLiveThresholdPercent=50 -XX:+AlwaysPreTouch -jar forge-1.16.5-36.2.39.jar nogui'"}

	ptrCmds := make([]*string, len(commands))
	for i, cmd := range commands {
		ptrCmds[i] = aws.String(cmd)
	}

	input := &ssm.SendCommandInput{
		InstanceIds:  []*string{&PixelmonInstanceID},
		DocumentName: &document,
		Parameters: map[string][]*string{
			"commands": ptrCmds,
		},
	}

	_, err = ssmSvc.SendCommand(input)
	if err != nil {
		return err.Error()
	}

	return Message[Starting]
}

func getSession() (*session.Session, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("PIXELMON_REGION")),
	})
	if err != nil {
		log.Printf("Error creating AWS session: %v", err)
		return nil, errors.New("Error creating AWS session")
	}

	return sess, nil
}

func getPixelmonServer(svc *ec2.EC2) (*ec2.DescribeInstancesOutput, error) {
	// Describe instances based on the "Name" tag
	input := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("tag:Name"),
				Values: []*string{aws.String(os.Getenv("PIXELMON_TAG"))},
			},
		},
	}

	return svc.DescribeInstances(input)
}
