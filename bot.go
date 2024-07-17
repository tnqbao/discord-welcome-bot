package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

var (
	configMutex sync.Mutex
	config      *Config
)

type Config struct {
	WelcomeChannels map[string]string `json:"welcome_channels"`
}

func loadConfig() (*Config, error) {
	file, err := os.Open("config.json")
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{WelcomeChannels: make(map[string]string)}, nil
		}
		return nil, err
	}
	defer file.Close()

	var config Config
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func saveConfig(config *Config) error {
	file, err := os.Create("config.json")
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(config)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	token := os.Getenv("DISCORD_BOT_TOKEN")
	if token == "" {
		log.Fatalf("No token provided. Set DISCORD_BOT_TOKEN environment variable.")
	}

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("Error creating Discord session: %v", err)
	}

	dg.AddHandler(func(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
		memberAdd(s, m)
	})
	dg.AddHandler(ready)

	err = dg.Open()
	if err != nil {
		log.Fatalf("Error opening Discord session: %v", err)
	}

	config, err = loadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	dg.AddHandler(interactionCreate)

	// Register the /welcome command
	_, err = dg.ApplicationCommandCreate(dg.State.User.ID, "", &discordgo.ApplicationCommand{
		Name:        "welcome",
		Description: "Set the welcome channel",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionChannel,
				Name:        "channel",
				Description: "The channel to send welcome messages",
				Required:    true,
			},
		},
	})
	if err != nil {
		log.Fatalf("Error creating command: %v", err)
	}

	fmt.Println("Bot is now running. Press CTRL+C to exit.")
	select {}
}

func ready(s *discordgo.Session, event *discordgo.Ready) {
	fmt.Println("Bot is ready!")
}

func memberAdd(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	configMutex.Lock()
	defer configMutex.Unlock()

	channelID, ok := config.WelcomeChannels[m.GuildID]
	if !ok {
		log.Printf("No welcome channel set for guild: %s", m.GuildID)
		return
	}

	message := fmt.Sprintf("Chào mừng <@%s> đến với máy chủ!", m.User.ID)
	_, err := s.ChannelMessageSend(channelID, message)
	if err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

func interactionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	if i.ApplicationCommandData().Name == "welcome" {
		options := i.ApplicationCommandData().Options
		var channelID string
		for _, option := range options {
			if option.Name == "channel" {
				channelID = option.ChannelValue(s).ID
			}
		}

		if channelID == "" {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Channel ID is required.",
				},
			})
			return
		}

		configMutex.Lock()
		config.WelcomeChannels[i.GuildID] = channelID
		configMutex.Unlock()

		err := saveConfig(config)
		if err != nil {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Error saving config: %v", err),
				},
			})
			return
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Welcome channel set to <#%s>", channelID),
			},
		})
	}
}
