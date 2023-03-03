package model

import (
	pq "github.com/emirpasic/gods/queues/priorityqueue"
	"github.com/emirpasic/gods/utils"
)

type QueryResult struct {
	result *pq.Queue // store *DocRank
}

func (qr *QueryResult) Add(doc_rank *DocRank) {
	qr.result.Enqueue(doc_rank)
}

func (qr *QueryResult) newQueryResult() {
	qr.result = pq.NewWith(byPriority)
}

func byPriority(a, b interface{}) int {
	priorityA := a.(*DocRank).score
	priorityB := b.(*DocRank).score
	return -utils.Float64Comparator(priorityA, priorityB) // "-" descending order
}

func NewQueryResult() *QueryResult {
	qr := QueryResult{}
	qr.newQueryResult()
	return &qr
}

func (qr *QueryResult) Empty() bool {
	return qr.result.Empty()
}

func (qr *QueryResult) Front() (*DocRank, bool) {
	result, ok := qr.result.Dequeue()
	return result.(*DocRank), ok
}

func (qr *QueryResult) Size() int {
	return qr.result.Size()
}
