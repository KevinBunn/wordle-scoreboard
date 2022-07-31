package main

import (
	db "WordleScoreboard/database"
	user "WordleScoreboard/user"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"log"
	"math"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"
)

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
		guessCountInt, _ := strconv.Atoi(guessCount)
		if strings.HasSuffix(subMessages[2], "*") {
			isOnHard = true
		}

		if wordleCount != 0 && (guessCount == "X" || (guessCountInt >= 1 && guessCountInt <= 6)) { //this should check if the entry is valid with these additional parameters
			/*
				The below should only be executed if it is decisively understood that the entry is valid
			*/

			score = getScore(guessCount, isOnHard)
			entryDayIndex := getDayIndexFromWordleNumber(wordleCount)
			entryWeekIndex := getWeekIndexFromWordleNumber(wordleCount)
			nowWeekIndex := getWeekIndexFromCurrentDay()

			err := db.UpdateUserScore(m.Author.ID, score, wordleCount)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Could not create user")
			} else {
				s.ChannelMessageSend(m.ChannelID, "Score for #"+strconv.Itoa(wordleCount)+": "+strconv.Itoa(score)+". Day Index: "+strconv.Itoa(entryDayIndex)+"; Week Index: "+strconv.Itoa(entryWeekIndex)+"; Real Week Index: "+strconv.Itoa(nowWeekIndex))
			}
		}
	}
}

func WeeklyReset() {
	// TODO: get all the users and create a list of User to easily iterate through them and their scores.
	// Example of how to initialize a new user,
	exampleUser := user.User{
		Id:                   "21354",
		FirstPlaceCount:      2,
		WeeklyScore:          3,
		MostRecentSubmission: 4,
		TotalAverage:         5.67,
		WeekDayScoreMap: map[string][]int{
			"currentWeek": {},
			"lastWeek":    {},
		},
	}
	updateUserList := []user.User{exampleUser}
	// TODO: Calculate Winner and update their first place count

	// TODO: pass in the user list, preferable with all of the data already changed here so we can
	// easily pass it into the firestore update function
	db.WeeklyReset(updateUserList)
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

func getDayIndexFromWordleNumber(wordleNumber int) int {
	/*387 is a Monday (0)
	387 % 7 = 2
	Therefore, the index is (n - 2) % 7
	where n = the wordle number
	*/

	dayIndex := (wordleNumber - 2) % 7

	return dayIndex
}

func getWeekIndexFromWordleNumber(wordleNumber int) int {
	weekIndex := (wordleNumber - 2) / 7 //does this automatically floor to int? It does in C#.....

	return weekIndex
}

func getWeekIndexFromCurrentDay() int {
	//week 55 is the week this wordle was made, Mon Jul 11
	mondayOfWeek55 := time.Date(2022, 07, 11, 0, 0, 0, 0, time.UTC)

	timeSinceMondayOfWeek55 := time.Now().Sub(mondayOfWeek55).Hours()

	return int(math.Floor(timeSinceMondayOfWeek55))/(24*7) + 55
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	db.StartFireBase()
	discord := startServer()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop

	discord.Close()
	db.CloseFireBase()
}
