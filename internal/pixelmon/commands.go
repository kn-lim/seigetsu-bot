package pixelmon

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/kn-lim/seigetsu-bot/internal/mcstatus"
)

func GetStatus() (string, error) {
	// Setup AWS Session
	cfg, err := getConfig()
	if err != nil {
		return "", err
	}

	// Get Pixelmon EC2 instance
	_, instance, err := getPixelmonServer(cfg)
	if err != nil {
		return "", err
	}

	// Get Pixelmon EC2 instance status
	if instance.State.Name == "running" {
		return Message[Online], nil
	}
	return Message[Offline], nil
}

func Start() error {
	log.Println("Starting Pixelmon EC2 instance...")

	// Setup AWS Session
	cfg, err := getConfig()
	if err != nil {
		return err
	}

	// Get Pixelmon EC2 instance
	client, instance, err := getPixelmonServer(cfg)
	if err != nil {
		return err
	}

	// Start Pixelmon EC2 instance
	if instance.State.Name == "running" {
		return errors.New(Message[Online])
	} else if instance.State.Name == "stopped" {
		input := &ec2.StartInstancesInput{
			InstanceIds: []string{*instance.InstanceId},
		}

		_, err := client.StartInstances(context.TODO(), input)
		if err != nil {
			log.Printf("Failed to start pixelmon: %v", err)
			return errors.New(Message[Err_Start])
		}
	}

	log.Println("Started Pixelmon EC2 instance")

	return nil
}

func Stop() error {
	log.Println("Stopping Pixelmon EC2 instance")

	// Setup AWS Session
	cfg, err := getConfig()
	if err != nil {
		return err
	}

	// Get Pixelmon EC2 instance
	client, instance, err := getPixelmonServer(cfg)
	if err != nil {
		return err
	}

	// Stop Pixelmon
	if instance.State.Name == "stopped" {
		return errors.New(Message[Offline])
	} else if instance.State.Name == "running" {
		input := &ec2.StopInstancesInput{
			InstanceIds: []string{*instance.InstanceId},
		}

		_, err := client.StopInstances(context.TODO(), input)
		if err != nil {
			log.Printf("Failed to stop pixelmon: %v", err)
			return errors.New(Message[Err_Stop])
		}
	}

	log.Println("Stopped Pixelmon EC2 instance")

	return nil
}

func StartPixelmon() error {
	log.Println("Starting Pixelmon service...")

	// Wait till Pixelmon EC2 instance is running
	for {
		msg, err := GetStatus()
		if err != nil {
			return err
		}

		if msg == Message[Online] {
			break
		}

		time.Sleep(delay * time.Second)
	}

	log.Println("Pixelmon EC2 instance is running")

	// Check if Pixelmon service is already running
	isOnline, _, err := mcstatus.GetMCStatus()
	if err != nil {
		return err
	}
	if isOnline {
		log.Printf("%v.%v is online", os.Getenv("PIXELMON_SUBDOMAIN"), os.Getenv("PIXELMON_DOMAIN"))
		return nil
	}

	cfg, err := getConfig()
	if err != nil {
		return err
	}

	// Create Pixelmon DNS Entry
	_, instance, err := getPixelmonServer(cfg)
	if err != nil {
		return err
	}
	createPixelmonDNSEntry(cfg, instance, os.Getenv("PIXELMON_HOSTED_ZONE_ID"), os.Getenv("PIXELMON_DOMAIN"), os.Getenv("PIXELMON_SUBDOMAIN"))

	log.Println("Sending command to Pixelmon EC2 instance...")

	// Send start command to Pixelmon EC2 instance
	client := ssm.NewFromConfig(cfg)
	documentName := "AWS-RunShellScript"
	params := map[string][]string{
		"commands": {"cd /opt/pixelmon/ && tmux new-session -d -s minecraft './start.sh'"},
	}
	input := &ssm.SendCommandInput{
		InstanceIds:  []string{os.Getenv("PIXELMON_INSTANCE_ID")},
		DocumentName: &documentName,
		Parameters:   params,
	}
	_, err = client.SendCommand(context.TODO(), input)
	if err != nil {
		return err
	}

	log.Println("Sent command to Pixelmon EC2 instance")

	for {
		isOnline, _, err := mcstatus.GetMCStatus()
		if err != nil {
			return err
		}

		if isOnline {
			log.Printf("%v.%v is online", os.Getenv("PIXELMON_SUBDOMAIN"), os.Getenv("PIXELMON_DOMAIN"))
			break
		}

		log.Println("Waiting for Pixelmon service to start...")
		time.Sleep(delay * time.Second)
	}

	log.Println("Started Pixelmon service")

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

		time.Sleep(delay * time.Second)
	}

	log.Println("Pixelmon EC2 instance is running")

	cfg, err := getConfig()
	if err != nil {
		return err
	}

	// Delete Pixelmon DNS Entry
	_, instance, err := getPixelmonServer(cfg)
	if err != nil {
		return err
	}
	deletePixelmonDNSEntry(cfg, instance, os.Getenv("PIXELMON_HOSTED_ZONE_ID"), os.Getenv("PIXELMON_DOMAIN"), os.Getenv("PIXELMON_SUBDOMAIN"))

	// Send start command to Pixelmon EC2 instance
	client := ssm.NewFromConfig(cfg)
	documentName := "AWS-RunShellScript"
	params := map[string][]string{
		"commands": {"mcrcon -H localhost -p " + os.Getenv("RCON_PASSWORD") + " \"stop\""},
	}
	input := &ssm.SendCommandInput{
		InstanceIds:  []string{os.Getenv("PIXELMON_INSTANCE_ID")},
		DocumentName: &documentName,
		Parameters:   params,
	}
	_, err = client.SendCommand(context.TODO(), input)
	if err != nil {
		return err
	}

	for {
		isOnline, _, err := mcstatus.GetMCStatus()
		if err != nil {
			return err
		}

		if !isOnline {
			log.Printf("%v.%v is offline", os.Getenv("PIXELMON_SUBDOMAIN"), os.Getenv("PIXELMON_DOMAIN"))
			time.Sleep(delay * time.Second)
			break
		}

		log.Println("Waiting for Pixelmon service to stop...")
		time.Sleep(delay * time.Second)
	}

	return nil
}
