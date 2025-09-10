package poll

type Poll struct {
	ID       int64          `json:"id"`
	Question string         `json:"question"`
	Options  []string       `json:"options"`
	Votes    map[string]int `json:"votes"`   // IP+pollID -> 1
	Results  map[string]int `json:"results"` // option -> count
}
