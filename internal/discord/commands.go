package discord

import (
	"log"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/kn-lim/seigetsu-bot/internal/mcstatus"
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
				err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: ":red_circle:   Error! Something went wrong with getting the required role IDs!",
					},
				})
				if err != nil {
					log.Printf("Error: %v", err)
				}

				return
			}

			switch i.ApplicationCommandData().Options[0].Name {
			case "status":
				// log.Println("/pixelmon status")

				// Get status
				msg, err := pixelmon.GetStatus()
				if err != nil {
					msg = err.Error()
				}

				err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: msg,
					},
				})
				if err != nil {
					log.Printf("Error: %v", err)
				}
			case "start":
				// log.Println("/pixelmon start")

				// Check if user has the required role to use command
				if !checkForMinecraftersRole(requiredRoles, s, i) {
					err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: "You don't have the required role to use this command!",
							Flags:   64,
						},
					})
					if err != nil {
						log.Printf("Error: %v", err)
					}

					return
				}

				err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: pixelmon.Message[pixelmon.Starting],
					},
				})
				if err != nil {
					log.Printf("Error: %v", err)
				}

				// Start Pixelmon EC2 Instance
				if err := pixelmon.Start(); err != nil {
					log.Printf("Error: %v", err)

					_, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
						Content: pixelmon.Message[pixelmon.Err_Start],
					})
					if err != nil {
						log.Fatalf("Error sending follow-up message: %v", err)
					}

					return
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

					return
				}

				_, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
					Content: pixelmon.Message[pixelmon.Online],
				})
				if err != nil {
					log.Fatalf("Error sending follow-up message: %v", err)
				}
			case "stop":
				// log.Println("/pixelmon stop")

				// Check if user has the required role to use command
				if !checkForMinecraftersRole(requiredRoles, s, i) {
					err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: "You don't have the required role to use this command!",
							Flags:   64,
						},
					})
					if err != nil {
						log.Printf("Error: %v", err)
					}

					return
				}

				err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: pixelmon.Message[pixelmon.Stopping],
					},
				})
				if err != nil {
					log.Printf("Error: %v", err)
				}

				// Stop Pixelmon service
				if err := pixelmon.StopPixelmon(); err != nil {
					log.Printf("Error: %v", err)

					_, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
						Content: pixelmon.Message[pixelmon.Err_Stop],
					})
					if err != nil {
						log.Fatalf("Error sending follow-up message: %v", err)
					}

					return
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

					return
				}

				_, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
					Content: pixelmon.Message[pixelmon.Offline],
				})
				if err != nil {
					log.Fatalf("Error sending follow-up message: %v", err)
				}
			case "whitelist":
				// log.Println("/pixelmon whitelist")

				// Check if user has the required role to use command
				if !checkForMinecraftersRole(requiredRoles, s, i) {
					err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: "You don't have the required role to use this command!",
							Flags:   64,
						},
					})
					if err != nil {
						log.Printf("Error: %v", err)
					}

					return
				}

				// Check if server is online
				isOnline, _, err := mcstatus.GetMCStatus()
				if err != nil {
					log.Printf("Error: %v", err)

					err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: pixelmon.Message[pixelmon.Err_Status],
							Flags:   64,
						},
					})
					if err != nil {
						log.Printf("Error: %v", err)
					}

					return
				}
				if !isOnline {
					err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: pixelmon.Message[pixelmon.Offline],
							Flags:   64,
						},
					})
					if err != nil {
						log.Printf("Error: %v", err)
					}

					return
				}

				// Get username to whitelist
				var username string
				options := i.ApplicationCommandData().Options
				for _, option := range options {
					if option.Name == "whitelist" {
						for _, subOption := range option.Options {
							if subOption.Name == "username" {
								username = subOption.StringValue()
								break
							}
						}
					}

					if username != "" {
						break
					}
				}

				err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: pixelmon.Message[pixelmon.Whitelist] + "`" + username + "`",
					},
				})
				if err != nil {
					log.Printf("Error: %v", err)
				}

				// Add name to whitelist
				if err := pixelmon.AddToWhitelist(username); err != nil {
					log.Printf("Error: %v", err)

					_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
						Content: pixelmon.Message[pixelmon.Err_Whitelist],
					})
					if err != nil {
						log.Fatalf("Error sending follow-up message: %v", err)
					}

					return
				}

				_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
					Content: pixelmon.Message[pixelmon.Success_Whitelist] + "`" + username + "`",
				})
				if err != nil {
					log.Fatalf("Error sending follow-up message: %v", err)
				}
			case "online":
				// log.Println("/pixelmon online")

				// Check if server is online
				isOnline, _, err := mcstatus.GetMCStatus()
				if err != nil {
					log.Printf("Error: %v", err)

					err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: pixelmon.Message[pixelmon.Err_Status],
							Flags:   64,
						},
					})
					if err != nil {
						log.Printf("Error: %v", err)
					}

					return
				}
				if !isOnline {
					err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: pixelmon.Message[pixelmon.Offline],
							Flags:   64,
						},
					})
					if err != nil {
						log.Printf("Error: %v", err)
					}

					return
				}

				num, err := pixelmon.GetNumberOfPlayers()
				if err != nil {
					log.Printf("Error: %v", err)

					err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: pixelmon.Message[pixelmon.Err_NumPlayers],
						},
					})
					if err != nil {
						log.Printf("Error: %v", err)
					}

					return
				}

				err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: pixelmon.Message[pixelmon.NumPlayers] + strconv.Itoa(num),
					},
				})
				if err != nil {
					log.Printf("Error: %v", err)
				}
			case "say":
				// log.Println("/pixelmon say")

				// Check if server is online
				isOnline, _, err := mcstatus.GetMCStatus()
				if err != nil {
					log.Printf("Error: %v", err)

					err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: pixelmon.Message[pixelmon.Err_Status],
							Flags:   64,
						},
					})
					if err != nil {
						log.Printf("Error: %v", err)
					}

					return
				}
				if !isOnline {
					err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: pixelmon.Message[pixelmon.Offline],
							Flags:   64,
						},
					})
					if err != nil {
						log.Printf("Error: %v", err)
					}

					return
				}

				// Get message to send
				var message string
				options := i.ApplicationCommandData().Options
				for _, option := range options {
					if option.Name == "say" {
						for _, subOption := range option.Options {
							if subOption.Name == "message" {
								message = subOption.StringValue()
								break
							}
						}
					}

					if message != "" {
						break
					}
				}

				err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: pixelmon.Message[pixelmon.SendingMessage] + "`" + message + "`",
					},
				})
				if err != nil {
					log.Printf("Error: %v", err)
				}

				// Send message
				if err := pixelmon.SendMessage(message); err != nil {
					log.Printf("Error: %v", err)

					_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
						Content: pixelmon.Message[pixelmon.Err_SendingMessage],
					})
					if err != nil {
						log.Fatalf("Error sending follow-up message: %v", err)
					}

					return
				}

				_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
					Content: pixelmon.Message[pixelmon.Success_SendingMessage] + "`" + message + "`",
				})
				if err != nil {
					log.Fatalf("Error sending follow-up message: %v", err)
				}
			}
		},
	}
)
