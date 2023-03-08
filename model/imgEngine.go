package model

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/emirpasic/gods/maps/hashmap"
	"github.com/emirpasic/gods/sets/treeset"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/yanyiwu/gojieba"
)

type ImgEngine struct {
	Imagine      *LStorage // img_id -> img_info
	IndexToIds   *LStorage // index(keyword) -> img_ids
	IdToKeywords *LStorage // img_id -> []keywords
	Tokenizer    *gojieba.Jieba
}

// 初始化ImgEngine
// 打开ImgEngine的leveldb
func (e *ImgEngine) Init() {
	e.Tokenizer = gojieba.NewJieba()
	e.Imagine = &LStorage{}
	err := e.Imagine.Open("LevelDB_img/img")
	if err != nil {
		log.Fatal(err)
	}
	e.IdToKeywords = &LStorage{}
	err = e.IdToKeywords.Open("LevelDB_img/id_keywords")
	if err != nil {
		log.Fatal(err)
	}
	e.IndexToIds = &LStorage{}
	err = e.IndexToIds.Open("LevelDB_img/index_ids")
	if err != nil {
		log.Fatal(err.Error())
	}
}

// 解析text并返回与其相关的结果
func (e *ImgEngine) Query(text string) *QueryResult {
	tokens := e.Tokenizer.CutForSearch(text, use_hmm)
	result := NewQueryResult()
	doc_record := treeset.NewWithStringComparator()
	for _, token := range tokens {
		ids, found := e.HandleIndex(token)
		if !found {
			continue
		}
		for _, id := range ids {
			if doc_record.Contains(id) {
				continue
			}
			doc_record.Add(id)
			score, _ := e.GetRank(id, tokens)
			result.Add(&DocRank{
				id:    id,
				score: score,
			})
		}
	}
	return result
}

func (e *ImgEngine) HandleIndex(index string) (ids []string, found bool) {
	if e.IndexToIds.IsClosed() {
		return nil, false
	}
	buf, err := e.IndexToIds.Get([]byte(index))
	if err == leveldb.ErrNotFound {
		return nil, false
	}
	ids = []string{}
	err = json.Unmarshal(buf, &ids)
	if err != nil {
		return nil, false
	}
	return ids, true
}

// 给指定id的文档 根据传入的tokens进行一次契合度评分
func (e *ImgEngine) GetRank(id string, tokens []string) (float64, error) {
	base := 1.0
	cnt := 1
	size := len(tokens)
	keys_for_img, err := e.GetKeywrdsById(id)
	doc_keys := hashmap.New()
	for _, v := range keys_for_img {
		doc_keys.Put(v.Word, v.Weight)
	}
	if err != nil {
		return 0, err
	}
	for i := 0; i < size; i++ {
		if val, find := doc_keys.Get(tokens[i]); find {
			weight := val.(float64)
			base += float64(size-i) * weight / 10
			cnt++
		}

	}
	return base * float64(cnt), nil
}

// 将文档存入 ImgEngine 的 imagineDB中
func (e *ImgEngine) AddImgs(img_path string, start_id int) (ok bool, add_cnt int, err error) {
	if e.Imagine.IsClosed() {
		return false, 0, fmt.Errorf("DB is closed")
	}
	file, err := os.Open(img_path)
	if err != nil {
		log.Fatal(err)
	}
	creader := csv.NewReader(file)
	creader.Read() //读取行首
	// 循环读取直到文件尾
	id := start_id
	for {
		record, err := creader.Read()
		if err != nil {
			break
		}
		img := Img{
			Url:         record[0],
			Description: record[1],
		}
		buf, err := json.Marshal(img)
		if err != nil {
			return false, id - start_id, err
		}
		err = e.Imagine.Put([]byte(strconv.Itoa(id)), buf)
		if err != nil {
			return false, id - start_id, err
		}
		id++
	}
	return true, id - start_id, nil
}

// 从ImgEngine的 imagineDB 中取出 id 对应的img
func (e *ImgEngine) GetImg(id string) (img *Img, err error) {
	if e.Imagine.IsClosed() {
		return nil, fmt.Errorf("DB is closed")
	}
	buf, err := e.Imagine.Get([]byte(id))
	if err != nil {
		return nil, err
	}
	img = &Img{}
	err = json.Unmarshal(buf, img)
	return
}

// 为ImgEngine的 idtokeywords 添加 id->[]keywords
// 会提取 img 中的前5(最多)个关键字存入leveldb中，采用 jieba的 extract(TF/IDF)
// 在addImg后为新添加的img添加关键字集合
func (e *ImgEngine) AddImgKeywords(id string) (ok bool, err error) {
	if e.IdToKeywords.IsClosed() {
		return false, fmt.Errorf("DB is closed")
	}
	img, err := e.GetImg(id)
	if err != nil {
		return false, err
	}
	s := e.Tokenizer.ExtractWithWeight(img.Description, 5)
	buf, err := json.Marshal(s)
	if err != nil {
		return false, err
	}
	err = e.IdToKeywords.Put([]byte(id), buf)
	if err != nil {
		return false, err
	}
	return true, nil
}

// 获取从img的描述提取的关键字
// TF/IDF 前5个
func (e *ImgEngine) GetKeywrdsById(id string) (keywords []gojieba.WordWeight, err error) {
	if e.IdToKeywords.IsClosed() {
		return nil, fmt.Errorf("DB is closed")
	}
	data, err := e.IdToKeywords.db.Get([]byte(id), nil)
	if err != nil {
		fmt.Printf("id: %v\n", id)
		fmt.Printf("err: %v\n", err)
		return nil, err
	}
	err = json.Unmarshal(data, &keywords)
	if err != nil {
		return nil, err
	}
	return
}

// 根据 索引关键字 获取其对应的文档ids
// index -> []id
func (e *ImgEngine) GetIdsByIndex(index string) (ids []string, err error) {
	if e.IndexToIds.IsClosed() {
		return nil, fmt.Errorf("DB is closed")
	}
	buf, err := e.IndexToIds.Get([]byte(index))
	if err == leveldb.ErrNotFound {
		return ids, nil
	}
	ids = []string{}
	err = json.Unmarshal(buf, &ids)
	if err != nil {
		return nil, err
	}
	return ids, err
}

// 将 index关键字 与 文档id相关联(将文档id加入到 关键字倒排索引的结果集中)
func (e *ImgEngine) ConnectIndexWithId(index string, id string) (ok bool, err error) {
	if e.IndexToIds.IsClosed() {
		return false, fmt.Errorf("DB is closed")
	}
	ids, err := e.GetIdsByIndex(index)
	if err != nil {
		return false, err
	}
	ids = append(ids, id)
	buf, err := json.Marshal(ids)
	if err != nil {
		return false, err
	}
	err = e.IndexToIds.Put([]byte(index), buf)
	if err != nil {
		return false, err
	}
	return true, err
}
