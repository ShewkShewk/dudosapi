package main

type Tournament struct {
	Id          int    `json:"id"`
	Date        string `json:"date"`
	Name        string `json:"name"`
	Loaded      bool   `json:"loaded"`
	UpdatedTime string `json:"updatedTime"`
}

type BallotResult string

const (
	WIN     BallotResult = "WIN"
	LOSS    BallotResult = "LOSS"
	BYE     BallotResult = "BYE"
	FFT     BallotResult = "FFT"
	UNKNOWN BallotResult = "UNKNOWN"
)

type Entry struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type Judge struct {
	Id       int    `json:"id"`
	PersonId int    `json:"personId"`
	Name     string `json:"name"`
	Started  bool   `json:"started"`
}

type Pairing struct {
	SectionId int           `json:"sectionId"`
	Flight    int           `json:"flight"`
	Room      *string       `json:"room"`
	AffEntry  *Entry        `json:"affEntry"`
	AffResult *BallotResult `json:"affResult"`
	NegEntry  *Entry        `json:"negEntry"`
	NegResult *BallotResult `json:"negResult"`
	Judges    []Judge       `json:"judges"`
}

type EventPairing struct {
	Name      string    `json:"name"`
	Number    int       `json:"number"`
	Flighted  bool      `json:"flighted"`
	StartTime string    `json:"startTime"`
	Pairings  []Pairing `json:"pairings"`
}

type TournamentPairings struct {
	Name          string         `json:"name"`
	UpdateTime    string         `json:"updateTime"`
	EventPairings []EventPairing `json:"eventPairings"`
}

type Summary struct {
	TournamentCount int64 `json:"tournamentCount"`
}
