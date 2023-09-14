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

// GetStatus returns whether the EC2 instance is online or offline
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

// Start turns on the EC2 instance
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

// Stop turns off the EC2 instance
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

// StartPixelmon turns on the Pixelmon Minecraft service
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
	if err := createPixelmonDNSEntry(cfg, instance, os.Getenv("PIXELMON_HOSTED_ZONE_ID"), os.Getenv("PIXELMON_DOMAIN"), os.Getenv("PIXELMON_SUBDOMAIN")); err != nil {
		return err
	}

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

	// Check if Minecraft service is online
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

	// Set online flag to true
	online = true

	return nil
}

// StopPixelmon turns off the Pixelmon Minecraft service
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
	if err := deletePixelmonDNSEntry(cfg, instance, os.Getenv("PIXELMON_HOSTED_ZONE_ID"), os.Getenv("PIXELMON_DOMAIN"), os.Getenv("PIXELMON_SUBDOMAIN")); err != nil {
		return err
	}

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

	// Set online flag to false
	online = false

	// Checks if Minecraft service is offline
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

// AddToWhitelist takes a username and runs the /whitelist add command
func AddToWhitelist(username string) error {
	cfg, err := getConfig()
	if err != nil {
		return err
	}

	// Send start command to Pixelmon EC2 instance
	client := ssm.NewFromConfig(cfg)
	documentName := "AWS-RunShellScript"
	params := map[string][]string{
		"commands": {"mcrcon -H localhost -p " + os.Getenv("RCON_PASSWORD") + " \"whitelist add " + username + "\""},
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

	return nil
}

// GetNumberOfPlayers gets the number of online players on the Minecraft server
func GetNumberOfPlayers() (int, error) {
	_, num, err := mcstatus.GetMCStatus()
	if err != nil {
		return 0, err
	}

	return num, nil
}

// SendMessage takes a message and runs the /say command
func SendMessage(msg string) error {
	cfg, err := getConfig()
	if err != nil {
		return err
	}

	// Send start command to Pixelmon EC2 instance
	client := ssm.NewFromConfig(cfg)
	documentName := "AWS-RunShellScript"
	params := map[string][]string{
		"commands": {"mcrcon -H localhost -p " + os.Getenv("RCON_PASSWORD") + " \"say " + msg + "\""},
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

	return nil
}
