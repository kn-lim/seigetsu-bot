package pixelmon

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/kn-lim/seigetsu-bot/internal/mcstatus"
)

var online = false

func CheckForOnlinePlayers(s *discordgo.Session) {
	var counter = 0

	for {
		if !online {
			// Pixelmon is offline, check again after the delay
			time.Sleep(delay * time.Minute)
			continue
		} else {
			_, num, err := mcstatus.GetMCStatus()
			if err != nil {
				log.Printf("Error with mcstatus: %v", err)
				time.Sleep(delay * time.Minute)
				continue
			}

			if num > 0 {
				// Pixelmon has online players, check again after the delay
				time.Sleep(delay * time.Minute)
			} else {
				// Pixelmon is online, but with no online players
				if counter == 0 {
					// First detection of no online players
					counter++
					time.Sleep(delay / 2 * time.Minute)
				} else if counter == 1 {
					// Second detection of no online players
					if err := StopPixelmon(); err != nil {
						log.Printf("Error stopping Pixelmon: %v", err)
					}

					// Send a message to indicate server shutting down
					s.ChannelMessageSend(os.Getenv("DISCORD_BOT_MSG_CHANNEL_ID"), ":red_square:   Shutting down Pixelmon due to no online players.")

					if err := StopPixelmon(); err != nil {
						s.ChannelMessageSend(os.Getenv("DISCORD_BOT_MSG_CHANNEL_ID"), fmt.Sprintf("Error with stopping Pixelmon: %v", err))
						log.Printf("Error with stopping Pixelmon: %v", err)
						online = false
						continue
					}

					// Send a message to indicate server shutting down
					s.ChannelMessageSend(os.Getenv("DISCORD_BOT_MSG_CHANNEL_ID"), Message[Offline])

					// Reset
					counter = 0
				}
			}
		}
	}
}
