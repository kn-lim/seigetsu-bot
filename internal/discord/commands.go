package discord

import (
	"log"
	"strconv"

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
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "whitelist",
					Description: "Adds a user to the whitelist of the Pixelmon server",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "username",
							Description: "Minecraft username to whitelist",
							Required:    true,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "online",
					Description: "List number of online players on the Pixelmon server",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "say",
					Description: "Sends a message to the Pixelmon server",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "message",
							Description: "Message to send to the Pixelmon server",
							Required:    true,
						},
					},
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
				// log.Println("/pixelmon status")

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
				// log.Println("/pixelmon start")

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

					_, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
						Content: pixelmon.Message[pixelmon.Err_Start],
					})
					if err != nil {
						log.Fatalf("Error sending follow-up message: %v", err)
					}
				}

				// Start Pixelmon service
				if err := pixelmon.StartPixelmon(); err != nil {
					log.Printf("Error: %v", err)

					_, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
						Content: pixelmon.Message[pixelmon.Err_Start],
					})
					if err != nil {
						log.Fatalf("Error sending follow-up message: %v", err)
					}
				}

				_, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
					Content: pixelmon.Message[pixelmon.Online],
				})
				if err != nil {
					log.Fatalf("Error sending follow-up message: %v", err)
				}
			case "stop":
				// log.Println("/pixelmon stop")

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

					_, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
						Content: pixelmon.Message[pixelmon.Err_Stop],
					})
					if err != nil {
						log.Fatalf("Error sending follow-up message: %v", err)
					}
				}

				// Stop Pixelmon EC2 Instance
				if err := pixelmon.Stop(); err != nil {
					log.Printf("Error: %v", err)

					_, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
						Content: pixelmon.Message[pixelmon.Err_Stop],
					})
					if err != nil {
						log.Fatalf("Error sending follow-up message: %v", err)
					}
				}

				_, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
					Content: pixelmon.Message[pixelmon.Offline],
				})
				if err != nil {
					log.Fatalf("Error sending follow-up message: %v", err)
				}
			case "whitelist":
				// log.Println("/pixelmon whitelist")

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

				// Check if server is online
				msg, err := pixelmon.GetStatus()
				if err != nil {
					log.Printf("Error: %v", err)

					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: pixelmon.Message[pixelmon.Err_Status],
						},
					})
				}
				if msg == pixelmon.Message[pixelmon.Offline] {
					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: pixelmon.Message[pixelmon.Offline],
						},
					})
				}

				// Get name to whitelist
				name := i.ApplicationCommandData().Options[0].StringValue()

				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: pixelmon.Message[pixelmon.Whitelist] + "`" + name + "`",
					},
				})

				// Add name to whitelist
				if err := pixelmon.AddToWhitelist(name); err != nil {
					log.Printf("Error: %v", err)

					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: pixelmon.Message[pixelmon.Err_Whitelist],
						},
					})
				}

				_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
					Content: pixelmon.Message[pixelmon.Success_Whitelist] + "`" + name + "`",
				})
				if err != nil {
					log.Fatalf("Error sending follow-up message: %v", err)
				}
			case "online":
				// log.Println("/pixelmon online")

				// Check if server is online
				msg, err := pixelmon.GetStatus()
				if err != nil {
					log.Printf("Error: %v", err)

					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: pixelmon.Message[pixelmon.Err_Status],
						},
					})
				}
				if msg == pixelmon.Message[pixelmon.Offline] {
					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: pixelmon.Message[pixelmon.Offline],
						},
					})
				}

				num, err := pixelmon.GetNumberOfPlayers()
				if err != nil {
					log.Printf("Error: %v", err)

					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: pixelmon.Message[pixelmon.Err_NumPlayers],
						},
					})
				}

				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: pixelmon.Message[pixelmon.NumPlayers] + strconv.Itoa(num),
					},
				})
			case "say":
				// log.Println("/pixelmon say")

				// Check if server is online
				msg, err := pixelmon.GetStatus()
				if err != nil {
					log.Printf("Error: %v", err)

					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: pixelmon.Message[pixelmon.Err_Status],
						},
					})
				}
				if msg == pixelmon.Message[pixelmon.Offline] {
					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: pixelmon.Message[pixelmon.Offline],
						},
					})
				}

				// Get message to send
				msg = i.ApplicationCommandData().Options[0].StringValue()

				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: pixelmon.Message[pixelmon.SendingMessage] + "`" + msg + "`",
					},
				})

				// Send message
				if err := pixelmon.SendMessage(msg); err != nil {
					log.Printf("Error: %v", err)

					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: pixelmon.Message[pixelmon.Err_SendingMessage],
						},
					})
				}

				_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
					Content: pixelmon.Message[pixelmon.Success_SendingMessage] + "`" + msg + "`",
				})
				if err != nil {
					log.Fatalf("Error sending follow-up message: %v", err)
				}
			}
		},
	}
)
