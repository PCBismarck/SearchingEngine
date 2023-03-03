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

func (d *DocRank) GetId() string {
	return d.id
}

func (d *DocRank) GetScore() float64 {
	return d.score
}
