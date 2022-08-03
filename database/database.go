package database

import (
	user "WordleScoreboard/user"
	"cloud.google.com/go/firestore"
	"context"
	firebase "firebase.google.com/go"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"log"
	"math/big"
	"strconv"
)

// Global variables that will be accessed in most/all functions
var ctx = context.Background()
var client *firestore.Client

// StartFireBase Initialize Firebase connection
func StartFireBase() {
	serviceAccount := option.WithCredentialsFile("./google-credentials.json")
	app, err := firebase.NewApp(ctx, nil, serviceAccount)
	if err != nil {
		log.Fatalln(err)
	}
	client, err = app.Firestore(ctx)
	if err != nil {
		log.Fatalln(err)
	}
}

func CloseFireBase() {
	err := client.Close()
	if err != nil {
		log.Fatalln(err)
	}
}

func UpdateUserScore(uid string, score int, wordleCount int, dayIndex int, weekIndex int) error {
	// Check if user already exists
	userSnapshot, err := GetUserSnapshot(uid)
	if err == iterator.Done {
		// No User found, we need to add
		_, _, err := client.Collection("Users").Add(ctx, map[string]interface{}{
			"id":                   uid,
			"weeklyScore":          score,
			"mostRecentSubmission": wordleCount,
			"totalAverage":         score,
			"scoreMap": map[string]map[string]int{
				strconv.Itoa(weekIndex): {strconv.Itoa(dayIndex): score},
			},
		})
		return err
	} else if err != nil {
		return err
	} else {
		// use the snapshot to update a user

		// create a user struct to help read/manipulate data
		var tempUser user.User
		err = userSnapshot.DataTo(&tempUser)
		if err != nil {
			return err
		}

		// update score map

		if len(tempUser.ScoreMap[strconv.Itoa(weekIndex)]) == 0 {
			// initialize the secondary map first
			tempUser.ScoreMap[strconv.Itoa(weekIndex)] = map[string]int{}
		}
		tempUser.ScoreMap[strconv.Itoa(weekIndex)][strconv.Itoa(dayIndex)] = score

		// calculate weekly score
		var weeklyScore int
		for _, val := range tempUser.ScoreMap[strconv.Itoa(weekIndex)] {
			weeklyScore += val
		}

		unRoundedAverageScore := float64(weeklyScore) / float64(len(tempUser.ScoreMap))
		// this is the best way I could find to get 2 decimals of precision
		averageScoreString := big.NewFloat(unRoundedAverageScore).Text('f', 2)
		averageScore, _ := strconv.ParseFloat(averageScoreString, 64)

		_, err := userSnapshot.Ref.Update(ctx, []firestore.Update{
			{
				Path:  "weeklyScore",
				Value: weeklyScore,
			},
			{
				Path:  "totalAverage",
				Value: averageScore,
			},
			{
				Path:  "mostRecentSubmission",
				Value: wordleCount,
			},
			{
				Path:  "scoreMap",
				Value: tempUser.ScoreMap,
			},
		})
		return err
	}
}

func WeeklyReset(users []user.User) {
	// TODO: Iterate through the user list, get the user snapshot and then call update
}

func GetUserSnapshot(uid string) (*firestore.DocumentSnapshot, error) {
	iter := client.Collection("Users").Limit(1).Where("id", "==", uid).Documents(ctx)

	snapshot, err := iter.Next()
	return snapshot, err
}

func GetAllUsers() []map[string]interface{} {
	var userList []map[string]interface{}
	iter := client.Collection("Users").Documents(ctx)
	for {
		snapshot, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			// do something
		}
		userList = append(userList, snapshot.Data())
	}
	return userList
}
