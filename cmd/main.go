package main

import (
	"muzikas"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

func main() {
	token := os.Getenv("TOKEN")
	if token == "" {
		panic("Please provide a token in ENV variable 'TOKEN'")
	}

	session, _ := discordgo.New("Bot " + token)
	muzikas := muzikas.NewMuzikasBot(session)
	muzikas.Start()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Cleanly close down the Discord session.
	session.Close()
}
