package main

import (
	"fmt"
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	token := os.Getenv("DISCORD_BOT_TOKEN")
	if token == "" {
		log.Fatalf("No token provided. Set DISCORD_BOT_TOKEN environment variable.")
	}

	channelID := os.Getenv("DISCORD_CHANNEL_ID")
	if channelID == "" {
		log.Fatalf("No channel ID provided. Set DISCORD_CHANNEL_ID environment variable.")
	}

	intents := discordgo.IntentsGuildMembers | discordgo.IntentsGuilds
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("Error creating Discord session: %v", err)
	}

	dg.Identify.Intents = intents

	dg.AddHandler(func(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
		memberAdd(s, m, channelID)
	})
	dg.AddHandler(ready)

	err = dg.Open()
	if err != nil {
		log.Fatalf("Error opening Discord session: %v", err)
	}

	fmt.Println("Bot is now running. Press CTRL+C to exit.")
	select {}
}

func ready(s *discordgo.Session, event *discordgo.Ready) {
	fmt.Println("Bot is ready!")
}

func memberAdd(s *discordgo.Session, m *discordgo.GuildMemberAdd, channelID string) {
	message := fmt.Sprintf("Chào mừng <@%s> đến với máy chủ!", m.User.ID)
	_, err := s.ChannelMessageSend(channelID, message)
	if err != nil {
		log.Printf("Error sending message: %v", err)
	}
}
