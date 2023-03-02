package test

import (
	"fmt"
	"testing"

	"github.com/PCBismarck/SearchingEngine/model"
)

func TestQuery(t *testing.T) {
	engine := model.Engine{}
	engine.Init()
	to_test := []string{"ChatGPT"}
	for _, text := range to_test {
		fmt.Printf("text: %v\n", text)
		re := engine.Query(text)
		for !re.Empty() {
			dr, b := re.Front()
			if b {
				fmt.Printf("dr: %#v\n", dr)
			}
		}
	}
	// d, err := engine.GetDoc("999")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Printf("d.Text: %v\n", d.Text)
}

func TestExtract(t *testing.T) {
	engine := model.Engine{}
	engine.Init()
	doc, _ := engine.GetDoc("999")
	fmt.Printf("engine.Tokenizer.ExtractWithWeight(doc.Text, 50): %v\n", engine.Tokenizer.ExtractWithWeight(doc.Text, 50))
}
