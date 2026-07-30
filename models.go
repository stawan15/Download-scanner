package main

type File struct {
	Name string `json:"name"`
	Size int64  `json:"size_bytes"`
}

type MovePlan struct {
	Source      string
	Destination string
}

type Finding struct {
	Name    string   `json:"name"`
	Score   int      `json:"score"`
	Reasons []string `json:"reasons"`
}

type OrganizationLog struct {
	Moves []MovePlan `json:"moves"`
}
