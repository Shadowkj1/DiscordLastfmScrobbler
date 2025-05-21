// DiscordLastfmScrobbler project main.go
package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
	"bufio"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/go-ini/ini"
	"github.com/shkh/lastfm-go/lastfm"
)

func Print(text string) {
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	fmt.Println(currentTime, " - ", text)
}

func createConfigFile() (string, error){
		//create a new config file and prompt the user to set it up
		reader := bufio.NewReader(os.Stdin)

		fmt.Println("Config file not found. Would you like to create one step by step or manually?\nReply with 'step' or 'manual'")
		response := ""

		//get the response from the user
		input, _ := reader.ReadString('\n')
		response = strings.TrimSpace(input)

		fmt.Scanln(&response)
		fmt.Println("This is the response: ", response)
		for response != "step" && response != "manual" {
			fmt.Println("Please enter 'step' or 'manual'")
			fmt.Scanln(&response)
		}

		if response == "step" {
			fmt.Println("Please enter your LastFM API key:")
			var apiKey string
			fmt.Scanln(&apiKey)
			fmt.Println("Please enter your LastFM username:")
			var username string
			fmt.Scanln(&username)
			fmt.Println("Please enter your Discord token:")
			var token string
			fmt.Scanln(&token)
			fmt.Println("Please enter the interval in seconds:")
			var interval int
			fmt.Scanln(&interval)
		} else if response == "manual" {
			fmt.Println("Please create a config.ini file in this same directory with the following format:")
			fmt.Println("[lastfm]")
			fmt.Println("api_key = <your_api_key>")
			fmt.Println("username = <your_username>")
			fmt.Println("interval = <interval_in_seconds>")
			fmt.Println("[discord]")
			fmt.Println("token = <your_discord_token>")
			//quit the program
			return "Closing program...", nil
		}
		response = ""
		return response, nil
}

func checkDiscordToken () string {
	token := ""



	return token
}
 

func scrobbler(quit chan struct{}) error {
	//check if the config file even exists
	fmt.Println("We are here")
	cfg, err := ini.Load("config.ini")
	fmt.Println("We are here again")
	fmt.Printf("This is what is gained from: %v\n, and also err %v\n", cfg, err)

	if cfg == nil{
		createConfigFile()
		} else {
			fmt.Println(err)
			return err
		}
	

	token := cfg.Section("discord").Key("token").String()
	apiKey := cfg.Section("lastfm").Key("api_key").String()
	username := cfg.Section("lastfm").Key("username").String()
	configInterval, err := cfg.Section("lastfm").Key("interval").Int()

	if err != nil {
		fmt.Println(err)
		return err
	}

	api := lastfm.New(apiKey, "")

	Print("Settings loaded: config.ini")

	dg, err := discordgo.New(token)
	if err != nil {
		fmt.Println("Discord error: ", err)
		return err
	}
	Print("Authorized to Discord")
	if err := dg.Open(); err != nil {
		fmt.Println("Discord error: ", err)
		return err
	}
	Print("Connected to Discord")

	interval := time.Duration(configInterval*1000) * time.Millisecond
	ticker := time.NewTicker(interval)
	var prevTrack = ""

	//For loop that constantly checks for the current track on LastFM
	// and updates the Discord status if it changes
	for {
		select {
		case <-quit:
			Print("Scrobbler stopping...")
			dg.Close()
			return nil
		case <-ticker.C:
			result, err := api.User.GetRecentTracks(lastfm.P{"limit": "1", "user": username})
			if err != nil {
				fmt.Println("LastFM error: ", err)
			} else {
				if len(result.Tracks) > 0 {
					currentTrack := result.Tracks[0]
					isNowPlaying, _ := strconv.ParseBool(currentTrack.NowPlaying)
					trackName := currentTrack.Artist.Name + " - " + currentTrack.Name
					if isNowPlaying && prevTrack != trackName {
						prevTrack = trackName
						statusData := discordgo.UpdateStatusData{
							Activities: []*discordgo.Activity{
								{
									Name:    prevTrack,
									Type:    discordgo.ActivityTypeListening,
									Details: "LAST.FM",
									State:   "DiscordLastfmScrobbler",
								},
							},
							AFK:    false,
							Status: "online",
						}
						if err := dg.UpdateStatusComplex(statusData); err != nil {
							fmt.Println("Discord error: ", err)
							return err
						}
						Print("Now playing: " + trackName)
					}
				}
			}
		}
	}
}

func main() {
    quit := make(chan struct{})
    
    // Run scrobbler unless there is an error
    go func() {
        if err := scrobbler(quit); err != nil {
            fmt.Println("Error in scrobbler:", err)
        }
    }()
    
    // Wait for user input to exit
    fmt.Println("Scrobbler is running. Press Enter to exit.")
    fmt.Scanln()
    
    // Signal scrobbler to quit
    close(quit)
    time.Sleep(1 * time.Second)
    fmt.Println("Program terminated.")
}
