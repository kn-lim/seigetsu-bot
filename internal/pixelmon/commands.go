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

func GetStatus() (string, error) {
	// Setup AWS Session
	sess, err := getSession()
	if err != nil {
		return "", err
	}

	// Get Pixelmon EC2 instance
	_, instance, err := getPixelmonServer(sess)
	if err != nil {
		return "", err
	}

	// Get Pixelmon EC2 instance status
	if *instance.State.Name == "running" {
		return Message[Online], nil
	}
	return Message[Offline], nil
}

func Start() error {
	log.Println("Starting Pixelmon EC2 instance")

	// Setup AWS Session
	sess, err := getSession()
	if err != nil {
		return err
	}

	// Get Pixelmon EC2 instance
	svc, instance, err := getPixelmonServer(sess)
	if err != nil {
		return err
	}

	// Start Pixelmon EC2 instance
	if *instance.State.Name == "running" {
		return errors.New(Message[Online])
	} else if *instance.State.Name == "stopped" {
		startInput := &ec2.StartInstancesInput{
			InstanceIds: []*string{instance.InstanceId},
		}

		_, err := svc.StartInstances(startInput)
		if err != nil {
			log.Printf("Failed to start pixelmon: %v", err)
			return errors.New(Message[Err_Start])
		}
	}

	return nil
}

func Stop() error {
	log.Println("Stoppping Pixelmon EC2 instance")

	// Setup AWS Session
	sess, err := getSession()
	if err != nil {
		return err
	}

	// Get Pixelmon EC2 instance
	svc, instance, err := getPixelmonServer(sess)
	if err != nil {
		return err
	}

	// Stop Pixelmon
	if *instance.State.Name == "stopped" {
		return errors.New(Message[Offline])
	} else if *instance.State.Name == "running" {
		stopInput := &ec2.StopInstancesInput{
			InstanceIds: []*string{instance.InstanceId},
		}

		_, err := svc.StopInstances(stopInput)
		if err != nil {
			log.Printf("Failed to stop pixelmon: %v", err)
			return errors.New(Message[Err_Stop])
		}
	}

	return nil
}

func StartPixelmon() error {
	log.Println("Starting Pixelmon service")

	// Wait till Pixelmon EC2 instance is running
	for {
		msg, err := GetStatus()
		if err != nil {
			return err
		}

		if msg == Message[Online] {
			break
		}

		time.Sleep(10 * time.Second)
	}

	log.Println("Pixelmon EC2 instance is running")

	// Send start command to Pixelmon EC2 instance
	sess, err := getSession()
	if err != nil {
		return err
	}
	svc := ssm.New(sess)
	documentName := "AWS-RunShellScript"
	params := map[string][]*string{
		"commands": {aws.String("touch /tmp/hello.txt && cd /opt/pixelmon/ && tmux new-session -d -s minecraft './start.sh'")},
	}
	input := &ssm.SendCommandInput{
		InstanceIds:  []*string{aws.String(os.Getenv("PIXELMON_INSTANCE_ID"))},
		DocumentName: &documentName,
		Parameters:   params,
	}
	_, err = svc.SendCommand(input)
	if err != nil {
		return err
	}

	log.Println("Finished sending command to Pixelmon EC2 instance")

	return nil
}

func StopPixelmon() error {
	log.Println("Stopping Pixelmon service")

	// Wait till Pixelmon EC2 instance is running
	for {
		msg, err := GetStatus()
		if err != nil {
			return err
		}

		if msg == Message[Online] {
			break
		}

		time.Sleep(10 * time.Second)
	}

	log.Println("Pixelmon EC2 instance is running")

	// Send start command to Pixelmon EC2 instance
	sess, err := getSession()
	if err != nil {
		return err
	}
	svc := ssm.New(sess)
	documentName := "AWS-RunShellScript"
	params := map[string][]*string{
		"commands": {aws.String("mcrcon -H localhost -p " + os.Getenv("RCON_PASSWORD") + " \"stop\"")},
	}
	input := &ssm.SendCommandInput{
		InstanceIds:  []*string{aws.String(os.Getenv("PIXELMON_INSTANCE_ID"))},
		DocumentName: &documentName,
		Parameters:   params,
	}
	_, err = svc.SendCommand(input)
	if err != nil {
		return err
	}

	return nil
}

func getSession() (*session.Session, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("PIXELMON_REGION")),
	})
	if err != nil {
		log.Printf("Error creating AWS session: %v", err)
		return nil, errors.New("error creating AWS session")
	}

	return sess, nil
}

func getPixelmonServer(sess *session.Session) (*ec2.EC2, *ec2.Instance, error) {
	svc := ec2.New(sess)

	input := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{
			aws.String(os.Getenv("PIXELMON_INSTANCE_ID")),
		},
	}

	result, err := svc.DescribeInstances(input)
	if err != nil {
		return nil, nil, errors.New(Message[Err_Status])
	}

	if len(result.Reservations) == 0 || len(result.Reservations[0].Instances) == 0 {
		return nil, nil, errors.New(Message[Not_Found])
	}

	return svc, result.Reservations[0].Instances[0], nil
}
