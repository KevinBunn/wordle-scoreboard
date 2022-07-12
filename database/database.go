package database

import (
	user "WordleScoreboard/user"
	"cloud.google.com/go/firestore"
	"context"
	firebase "firebase.google.com/go"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"log"
)

// Global variables that will be accessed in most/all functions
var ctx = context.Background()
var client *firestore.Client

// Initialize Firebase connection
func StartFireBase() {
	serviceAccount := option.WithCredentialsFile("./service-account.json")
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

func UpdateUserScore(uid string, score int, wordleCount int) error {
	// Check if user already exists
	userSnapshot, err := GetUserSnapshot(uid)
	if err == iterator.Done {
		// No User found, we need to add
		_, _, err := client.Collection("Users").Add(ctx, map[string]interface{}{
			"id":                   uid,
			"weeklyScore":          score,
			"mostRecentSubmission": wordleCount,
			"totalAverage":         score,
			"weekDayScoreMap": map[string][]int{
				"currentWeek": {score}, // TODO: we need to address how many days we are into the week here
			},
		})
		return err
	} else if err != nil {
		return err
	} else {
		// use the snapshot to update a user
		// TODO: We need to grab the existing user week to append a new score onto it
		// TODO: calculate the new weekly score with the new array
		// TODO: average the scores together
		_, err := userSnapshot.Ref.Update(ctx, []firestore.Update{
			{
				Path:  "weeklyScore",
				Value: score, // TODO: make this the sum
			},
			{
				Path:  "totalAverage",
				Value: score, // TODO: make this the average
			},
			{
				Path:  "mostRecentSubmission",
				Value: wordleCount,
			},
			{
				Path: "weekDayScoreMap",
				Value: map[string][]int{
					"currentWeek": {score}, // we'll actually need to use append eventually
				},
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
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			// do something
		}
		userList = append(userList, doc.Data())
	}
	return userList
}
