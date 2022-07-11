package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"log"
	"os"
	"os/signal"
)

// It might be relevant to store the version, it might not
const Version = "v0.0.0-alpha"

func startServer() *discordgo.Session {
	discord, _ := discordgo.New("Bot " + os.Getenv("BOT_TOKEN"))

	discord.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		fmt.Println("Bot is ready")
	})
	discord.AddHandler(onMessage)

	err := discord.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	return discord
}

func onMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	prefix := os.Getenv("PREFIX")
	if m.Content == prefix+"hello" {
		s.ChannelMessageSend(m.ChannelID, "Waddup")
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	discord := startServer()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop

	discord.Close()
}
