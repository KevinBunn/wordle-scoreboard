package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"log"
	"math"
	"os"
	"os/signal"
	"strconv"
	"strings"
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
	/*prefix := os.Getenv("PREFIX")
	if m.Content == prefix+"hello" {
		s.ChannelMessageSend(m.ChannelID, "Waddup")
	}*/

	subMessages := strings.Fields(m.Content)

	wordleCount := 0
	guessCount := ""
	isOnHard := false
	score := 0

	if subMessages[0] == "Wordle" && len(subMessages) >= 5 && !m.Author.Bot { //presumably is a wordle score
		wordleCount, _ = strconv.Atoi(subMessages[1])
		scoreSplit := strings.Split(subMessages[2], "/")
		guessCount = scoreSplit[0]
		if strings.HasSuffix(subMessages[2], "*") {
			isOnHard = true
		}

		score = getScore(guessCount, isOnHard)

		s.ChannelMessageSend(m.ChannelID, "Score for #"+strconv.Itoa(wordleCount)+": "+strconv.Itoa(score))
	}
}

/*
SCORE METHOD:
Where x = number of guesses, 1 <= x <= n
and n = total guesses possible (6)
and m = * (on hard) ? .4 : 0 (BONUS IF USER IS ON HARD MODE)

score y = floor( (n - x) ^ (1 + m) + 1)

DESMOS NOTATION:
y\ =\ \operatorname{floor}\left(\left(6-x\right)^{\left(1+.4\right)}\ +\ 1\right)\ \left\{1\ \le x\ \le6\right\}
*/

func getScore(guessCount string, isOnHard bool) int {
	score := 0
	scoreMultiplier := 0.0
	guessLimit := 6

	if isOnHard {
		scoreMultiplier = 0.4 //upset that golang doesn't support ternary one-liners
	}

	if guessCount == "X" {
		score = 1
	} else {
		guessCountInt, _ := strconv.Atoi(guessCount)

		if guessCountInt >= 1 && guessCountInt <= guessLimit {
			score = int(math.Floor(math.Pow(float64(guessLimit-guessCountInt), 1+scoreMultiplier) + 1))
		}
	}

	return score
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
