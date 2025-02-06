package models

type (
	Pagination struct {
		Links struct {
			First    string `json:"first"`
			Previous string `json:"previous"`
			Current  string `json:"current"`
			Next     string `json:"next"`
		} `json:"links"`
		Info struct {
			PerPage int `json:"per_page"`
			Pages   int `json:"pages"`
			Total   int `json:"total"`
		} `json:"info"`
	}
)

const (
	OldTestament = "OT"
	NewTestament = "NT"
)

var (
	MapPerPageRowsCount = map[int]int{1: 1, 10: 1, 25: 1, 50: 1, 100: 1}
	PerPageRowsCount    = []int{1, 10, 25, 50, 100}
)
