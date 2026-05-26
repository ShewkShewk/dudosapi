package sqlc

type AssociatedJudge struct {
	Id        int    `json:"id"`
	PersonId  int    `json:"personId"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Started   bool   `json:"started"`
}

type TeamBallot struct {
	Side   string `json:"side"`
	Result string `json:"result"`
	Judge  int    `json:"judge"`
}
