package user

type User struct {
	Id                   string
	FirstPlaceCount      int
	WeeklyScore          int
	MostRecentSubmission int
	TotalAverage         float32
	ScoreMap             map[string]string // a map of maps with a number value
}
