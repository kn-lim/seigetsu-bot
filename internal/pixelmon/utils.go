package pixelmon

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	route53Types "github.com/aws/aws-sdk-go-v2/service/route53/types"
)

const (
	Online = iota
	Offline
	Not_Found
	Starting
	Stopping
	Err_Status
	Err_Start
	Err_Stop
)

const (
	statusURL            = "https://api.mcstatus.io/v2/status/java/pixelmon.knlim.dev"
	delay                = 30
	MinecraftersRoleName = "Minecrafters"
)

var (
	RequiredRoleNames = []string{
		MinecraftersRoleName,
	}

	Message = []string{
		":green_circle: Pixelmon is online :green_circle:",
		":red_circle: Pixelmon is offline :red_circle:",
		":grey_exclamation: No Pixelmon server was found :grey_exclamation:",
		":green_square: Starting the Pixelmon server :green_square:",
		":red_square: Stopping the Pixelmon server :red_square:",
		":exclamation: Error checking Pixelmon's status :exclamation:",
		":exclamation: Failed to start the Pixelmon server :exclamation:",
		":exclamation: Failed to stop the Pixelmon server :exclamation:",
	}
)

func getConfig() (aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(os.Getenv("PIXELMON_REGION")))
	if err != nil {
		log.Printf("Error creating AWS config: %v", err)
		return aws.Config{}, errors.New("error creating AWS config")
	}

	return cfg, nil
}

func getPixelmonServer(cfg aws.Config) (*ec2.Client, ec2Types.Instance, error) {
	client := ec2.NewFromConfig(cfg)

	input := &ec2.DescribeInstancesInput{
		InstanceIds: []string{
			os.Getenv("PIXELMON_INSTANCE_ID"),
		},
	}

	result, err := client.DescribeInstances(context.TODO(), input)
	if err != nil {
		return nil, ec2Types.Instance{}, errors.New(Message[Err_Status])
	}

	if len(result.Reservations) == 0 || len(result.Reservations[0].Instances) == 0 {
		return nil, ec2Types.Instance{}, errors.New(Message[Not_Found])
	}

	return client, result.Reservations[0].Instances[0], nil
}

func createPixelmonDNSEntry(cfg aws.Config, instance ec2Types.Instance, zoneID string, domain string, subdomain string) error {
	publicIP := instance.PublicIpAddress

	client := route53.NewFromConfig(cfg)

	fqdn := fmt.Sprintf("%s.%s", subdomain, domain)
	log.Printf("Creating A record of %v to %v", *publicIP, fqdn)

	_, err := client.ChangeResourceRecordSets(context.TODO(), &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: &zoneID,
		ChangeBatch: &route53Types.ChangeBatch{
			Changes: []route53Types.Change{
				{
					Action: route53Types.ChangeActionUpsert,
					ResourceRecordSet: &route53Types.ResourceRecordSet{
						Name: &fqdn,
						Type: route53Types.RRTypeA,
						TTL:  aws.Int64(300),
						ResourceRecords: []route53Types.ResourceRecord{
							{
								Value: publicIP,
							},
						},
					},
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create A record: %v", err)
	}

	log.Printf("Created A record of %v to %v.%v", *publicIP, subdomain, domain)

	return nil
}

func deletePixelmonDNSEntry(cfg aws.Config, instance ec2Types.Instance, zoneID string, domain string, subdomain string) error {
	publicIP := instance.PublicIpAddress

	client := route53.NewFromConfig(cfg)

	fqdn := fmt.Sprintf("%s.%s", subdomain, domain)
	log.Printf("Deleting A record of %v to %v", *publicIP, fqdn)

	_, err := client.ChangeResourceRecordSets(context.TODO(), &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: &zoneID,
		ChangeBatch: &route53Types.ChangeBatch{
			Changes: []route53Types.Change{
				{
					Action: route53Types.ChangeActionDelete,
					ResourceRecordSet: &route53Types.ResourceRecordSet{
						Name: &fqdn,
						Type: route53Types.RRTypeA,
						TTL:  aws.Int64(300),
						ResourceRecords: []route53Types.ResourceRecord{
							{
								Value: publicIP,
							},
						},
					},
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create A record: %v", err)
	}

	log.Printf("Deleted A record of %v to %v.%v", *publicIP, subdomain, domain)

	return nil
}
