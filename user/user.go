package user

type User struct {
	Id                   string
	FirstPlaceCount      int
	WeeklyScore          int
	MostRecentSubmission int
	TotalAverage         float32
	ScoreMap             map[string]int // a map with a string key and an integer array value
}