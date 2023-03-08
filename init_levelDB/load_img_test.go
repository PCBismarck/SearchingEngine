package initleveldb

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"testing"

	"github.com/PCBismarck/SearchingEngine/model"
)

func AddImgs(db *model.ImgEngine, start_id int) (add_cnt int) {
	_, add_cnt, err := db.AddImgs("test.csv", start_id)
	if err != nil {
		log.Fatal(err)
	}
	return
}

func AddImgKeywors(db *model.ImgEngine, start_id int, add_cnt int) {
	for i := 0; i < add_cnt; i++ {
		_, err := db.AddImgKeywords(strconv.Itoa(start_id + i))
		if err != nil {
			log.Fatal(err)
		}
	}
}

func AddIndexes_img(db *model.ImgEngine, start_id int, add_cnt int) {
	for i := 0; i < add_cnt; i++ {
		id := strconv.Itoa(start_id + i)
		keywords, err := db.GetKeywrdsById(id)
		if err != nil {
			fmt.Println("GetKeywrdsById(id) falt")
			log.Fatal(err)
		}
		for _, v := range keywords {
			_, err := db.ConnectIndexWithId(v.Word, id)
			if err != nil {
				log.Fatal(err)
				fmt.Println("ConnectIndexWithId falt")
			}
		}
	}
}

func InitLevelDBImg(db *model.ImgEngine) {
	start_id := 0
	add_cnt := AddImgs(db, start_id)
	fmt.Printf("add_cnt: %v\n", add_cnt)
	AddImgKeywors(db, start_id, add_cnt)
	AddIndexes_img(db, start_id, add_cnt)
}

func TestLevelDBImg(t *testing.T) {
	DB := model.ImgEngine{}
	DB.Init()
	// InitLevelDBImg(&DB)
	i, err := DB.GetImg("0")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("i(: %#v\n", i)
	ww := DB.Tokenizer.ExtractWithWeight("今年能跑赢96不?备战坦克两项俄军开始选拔参赛队员", 5)
	fmt.Printf("ww: %#v\n", ww)
	val, err := DB.IndexToIds.Get([]byte("参赛"))
	var ids []string
	json.Unmarshal(val, &ids)
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Printf("val: %#v\n", val)
	fmt.Printf("ids: %#v\n", ids)
	// val2, _ := DB.IdToKeywords.Get([]byte("0"))
	// var res []gojieba.WordWeight
	// err = json.Unmarshal(val2, &res)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Printf("res: %#v\n", res)
}
