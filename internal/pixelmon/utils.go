package pixelmon

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

const statusURL = "https://api.mcstatus.io/v2/status/java/pixelmon.knlim.dev"

var (
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
