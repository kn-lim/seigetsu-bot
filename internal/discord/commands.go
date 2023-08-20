package discord

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/kn-lim/seigetsu-bot/internal/pixelmon"
)

var (
	Commands = []*discordgo.ApplicationCommand{
		{
			Name:        "pixelmon",
			Description: "Pixelmon command",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "status",
					Description: "Get the status of the Pixelmon server",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "start",
					Description: "Starts the Pixelmon server",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "stop",
					Description: "Stops the Pixelmon server",
				},
			},
		},
	}

	CommandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"pixelmon": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			requiredRoles, err := getRequiredRoleIDs(s, i)
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Error! Something went wrong!",
					},
				})
			}

			switch i.ApplicationCommandData().Options[0].Name {
			case "status":
				log.Println("/pixelmon status")

				msg, err := pixelmon.GetStatus()
				if err != nil {
					msg = err.Error()
				}

				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: msg,
					},
				})
			case "start":
				log.Println("/pixelmon start")

				if !checkForMinecraftersRole(requiredRoles, s, i) {
					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: "You don't have the required role to use this command!",
							Flags:   64,
						},
					})
					return
				}

				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: pixelmon.Message[pixelmon.Starting],
					},
				})

				// Start Pixelmon EC2 Instance
				if err := pixelmon.Start(); err != nil {
					log.Printf("Error: %v", err)
				}

				// Start Pixelmon service
				if err := pixelmon.StartPixelmon(); err != nil {
					log.Printf("Error: %v", err)
				}

				_, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
					Content: pixelmon.Message[pixelmon.Online],
				})
				if err != nil {
					log.Fatalf("Error sending follow-up message: %v", err)
				}
			case "stop":
				log.Println("/pixelmon stop")

				if !checkForMinecraftersRole(requiredRoles, s, i) {
					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: "You don't have the required role to use this command!",
							Flags:   64,
						},
					})
					return
				}

				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: pixelmon.Message[pixelmon.Stopping],
					},
				})

				// Stop Pixelmon service
				if err := pixelmon.StopPixelmon(); err != nil {
					log.Printf("Error: %v", err)
				}

				// Stop Pixelmon EC2 Instance
				if err := pixelmon.Stop(); err != nil {
					log.Printf("Error: %v", err)
				}

				_, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
					Content: pixelmon.Message[pixelmon.Offline],
				})
				if err != nil {
					log.Fatalf("Error sending follow-up message: %v", err)
				}
			}
		},
	}
)
