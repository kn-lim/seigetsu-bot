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
			switch i.ApplicationCommandData().Options[0].Name {
			case "status":
				log.Println("/pixelmon status")

				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: pixelmon.GetStatus(),
					},
				})
			case "start":
				log.Println("/pixelmon start")

				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: pixelmon.Start(),
					},
				})
			case "stop":
				log.Println("/pixelmon stop")

				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: pixelmon.Stop(),
					},
				})
			}
		},
	}
)
