package model

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/emirpasic/gods/maps/hashmap"
	"github.com/emirpasic/gods/sets/treeset"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/yanyiwu/gojieba"
)

const use_hmm = true

type Engine struct {
	Documents    *LStorage // doc_id -> doc_info
	IndexToIds   *LStorage // index(keyword) -> []doc_ids
	IdToKeywords *LStorage // doc_id -> []keywords
	Tokenizer    *gojieba.Jieba
}

// 初始化engine
// 打开engine的leveldb
func (e *Engine) Init() {
	e.Tokenizer = gojieba.NewJieba()
	e.Documents = &LStorage{}
	err := e.Documents.Open("LevelDB/doc")
	if err != nil {
		log.Fatal(err)
	}
	e.IdToKeywords = &LStorage{}
	err = e.IdToKeywords.Open("LevelDB/id_keywords")
	if err != nil {
		log.Fatal(err)
	}
	e.IndexToIds = &LStorage{}
	err = e.IndexToIds.Open("LevelDB/index_ids")
	if err != nil {
		log.Fatal(err.Error())
	}
}

// 解析text并返回与其相关的结果
func (e *Engine) Query(text string) *QueryResult {
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

func (e *Engine) HandleIndex(index string) (ids []string, found bool) {
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
func (e *Engine) GetRank(id string, tokens []string) (float64, error) {
	base := 1.0
	cnt := 1
	size := len(tokens)
	keys_for_doc, err := e.GetKeywrdsById(id)
	doc_keys := hashmap.New()
	for _, v := range keys_for_doc {
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

// 将文档存入 engine 的 documentsDB中
func (e *Engine) AddDoc(doc_path string, id string) (ok bool, err error) {
	if e.Documents.IsClosed() {
		return false, fmt.Errorf("DB is closed")
	}
	file, err := os.Open(doc_path)
	if err != nil {
		log.Fatal(err)
	}
	r := bufio.NewReader(file)
	buf, _ := r.ReadBytes('\n')
	link := string(buf)
	buf, _ = r.ReadBytes('\n')
	title := string(buf)
	buf, _ = os.ReadFile(doc_path)
	date := string(buf)
	buf, _ = os.ReadFile(doc_path)
	text := string(buf)
	doc := Doc{
		Link:  link,
		Title: title[7:],
		Text:  text,
		Date:  date,
	}
	buf, err = json.Marshal(doc)
	if err != nil {
		return false, err
	}
	err = e.Documents.Put([]byte(id), buf)
	if err != nil {
		return false, err
	}
	return true, nil
}

// 从engine的 documentsDB 中取出 id 对应的文档
func (e *Engine) GetDoc(id string) (doc *Doc, err error) {
	if e.Documents.IsClosed() {
		return nil, fmt.Errorf("DB is closed")
	}
	buf, err := e.Documents.Get([]byte(id))
	if err != nil {
		return nil, err
	}
	doc = &Doc{}
	err = json.Unmarshal(buf, doc)
	return
}

// 为engine的 idtokeywords 添加 id->[]keywords
// 会提取 doc 中的前30个关键字和weight存入leveldb中，采用 jieba的 extractwithweight(TF/IDF)
// 在addDoc后为新添加的doc添加关键字集合
func (db *Engine) AddDocKeywords(id string) (ok bool, err error) {
	if db.IdToKeywords.IsClosed() {
		return false, fmt.Errorf("DB is closed")
	}
	doc, err := db.GetDoc(id)
	if err != nil {
		return false, err
	}
	s := db.Tokenizer.ExtractWithWeight(doc.Text, 30)
	buf, err := json.Marshal(s)
	if err != nil {
		return false, err
	}
	err = db.IdToKeywords.Put([]byte(id), buf)
	if err != nil {
		return false, err
	}
	return true, nil
}

// 获取从文档提取的关键字
// TF/IDF 前20个
func (e *Engine) GetKeywrdsById(id string) (keywords []gojieba.WordWeight, err error) {
	if e.IdToKeywords.IsClosed() {
		return nil, fmt.Errorf("DB is closed")
	}
	data, err := e.IdToKeywords.db.Get([]byte(id), nil)
	if err != nil {
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
func (e *Engine) GetIdsByIndex(index string) (ids []string, err error) {
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
func (e *Engine) ConnectIndexWithId(index string, id string) (ok bool, err error) {
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
