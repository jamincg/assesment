package model

import (
	"container/list"
	"time"
)

// CacheItem - used by the priority queues and linked list
type CacheItem struct {
	Key        string
	Value      interface{}
	Priority   int
	ExpireTime time.Time
}

type PriorityQueueItem struct {
	Item  *CacheItem
	Index int
}

// PriorityQueue - items ordered based on priority
type PriorityQueue []*PriorityQueueItem

func (pq PriorityQueue) Len() int           { return len(pq) }
func (pq PriorityQueue) Less(i, j int) bool { return pq[i].Item.Priority < pq[j].Item.Priority }
func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].Index = i
	pq[j].Index = j
}

// Push item into priority queue
func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*PriorityQueueItem)
	item.Index = n
	*pq = append(*pq, item)
}

// Pop the first item from the priority queue
func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[:n-1]
	return item
}

// PQueueForExpiry - is a second PriorityQueue and items are ordered based on expiry
type PQueueForExpiry []*PriorityQueueItem

func (pqe PQueueForExpiry) Len() int { return len(pqe) }
func (pqe PQueueForExpiry) Less(i, j int) bool {
	return pqe[i].Item.ExpireTime.Before(pqe[j].Item.ExpireTime)
}
func (pqe PQueueForExpiry) Swap(i, j int) {
	pqe[i], pqe[j] = pqe[j], pqe[i]
	pqe[i].Index = i
	pqe[j].Index = j
}

// Push item into "expiry priority queue"
func (pqe *PQueueForExpiry) Push(x interface{}) {
	n := len(*pqe)
	item := x.(*PriorityQueueItem)
	item.Index = n
	*pqe = append(*pqe, item)
}

// Pop the first item from the "expiry priority queue"
func (pqe *PQueueForExpiry) Pop() interface{} {
	old := *pqe
	n := len(old)
	item := old[n-1]
	*pqe = old[:n-1]
	return item
}

// PriorityExpiryCache
type PriorityExpiryCache struct {
	MaxItems        int
	ItemMap         map[string]*list.Element
	LRUList         *list.List
	PriorityQueue   PriorityQueue
	PQueueForExpiry PQueueForExpiry
}
