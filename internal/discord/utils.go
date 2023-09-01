package discord

import (
	"github.com/bwmarrin/discordgo"
	"github.com/kn-lim/seigetsu-bot/internal/pixelmon"
)

// getRequiredRoleIDs gets the role IDs of roles required for users to run certain commands
func getRequiredRoleIDs(s *discordgo.Session, i *discordgo.InteractionCreate) (map[string]string, error) {
	if len(pixelmon.RequiredRoleNames) == 0 {
		return nil, nil
	}

	// Fetch all roles of the guild
	roles, err := s.GuildRoles(i.GuildID)
	if err != nil {
		return nil, err
	}

	requiredRoles := make(map[string]string)
	for _, role := range roles {
		for _, name := range pixelmon.RequiredRoleNames {
			if role.Name == name {
				requiredRoles[name] = role.ID
			}
		}
	}

	return requiredRoles, nil
}

// checkForMinecraftersRole checks to see if the user has the Minecrafters role
func checkForMinecraftersRole(requiredRoles map[string]string, s *discordgo.Session, i *discordgo.InteractionCreate) bool {
	hasRole := false
	for _, roleID := range i.Member.Roles {
		if roleID == requiredRoles[pixelmon.MinecraftersRoleName] {
			hasRole = true
			break
		}
	}

	return hasRole
}
