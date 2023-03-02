package model

type Doc struct {
	Link  string
	Title string
	Text  string
}

type DocRank struct {
	id    string
	score float64
}
