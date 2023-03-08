package initleveldb

import (
	"fmt"
	"log"
	"strconv"
	"testing"

	"github.com/PCBismarck/SearchingEngine/model"
)

const start, end = 0, 1000

// const doc_base_name = "doc/doc_"

// 将 doc/doc_x 的文档加入levedb
// 文档格式：1. 第一行是link 2. 第二行是title 3. 第三行往后开始是文档内容
func AddDocuments(db *model.Engine) {
	base := "doc/doc_"
	for i := start; i < end; i++ {
		db.AddDoc(fmt.Sprintf(base+"%v.txt", i), strconv.Itoa(i))
	}
}

func AddDocKeywors(db *model.Engine) {
	for i := start; i < end; i++ {
		_, err := db.AddDocKeywords(strconv.Itoa(i))
		if err != nil {
			log.Fatal(err)
		}
	}
}

func AddIndexes(db *model.Engine) {
	for i := start; i < end; i++ {
		id := strconv.Itoa(i)
		keywords, err := db.GetKeywrdsById(id)
		if err != nil {
			log.Fatal(err)
		}
		for _, v := range keywords {
			_, err := db.ConnectIndexWithId(v.Word, id)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func InitLevelDB(db *model.Engine) {
	AddDocuments(db)
	AddDocKeywors(db)
	AddIndexes(db)
}

func TestLevelDB(t *testing.T) {
	DB := model.Engine{}
	DB.Init()
	InitLevelDB(&DB)
}
