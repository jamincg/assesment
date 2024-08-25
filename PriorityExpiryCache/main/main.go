package main

import (
	"PriorityExpiryCache/model"
	"container/heap"
	"container/list"
	"fmt"
	"time"
)

type PriorityExpiryCache model.PriorityExpiryCache

// NewPriorityExpiryCache - constructor
func NewPriorityExpiryCache(maxItem int) *PriorityExpiryCache {
	return &PriorityExpiryCache{
		MaxItems:        maxItem,
		ItemMap:         make(map[string]*list.Element),
		LRUList:         list.New(),
		PriorityQueue:   model.PriorityQueue{},
		PQueueForExpiry: model.PQueueForExpiry{},
	}
}

// SetMaxItems update the cache capacity and evict items if needed
func (pe *PriorityExpiryCache) SetMaxItems(maxItem int) {
	pe.MaxItems = maxItem
	pe.evictItem()
}

// Add item to the cache
// Evict an item if the Cache has reached capacity
func (pe *PriorityExpiryCache) Set(key string, value interface{}, priority int, expire time.Duration) {

	expiry := time.Now().Add(expire)
	// if key exists, move it to the front of the linked list denoting that it was recently used
	if element, found := pe.ItemMap[key]; found {
		item := element.Value.(*model.CacheItem)
		item.Value = value
		item.Priority = priority
		item.ExpireTime = expiry

		pe.update(key) // update the priority queues
		pe.LRUList.MoveToFront(element)
		return
	}

	// add new Item
	newItem := &model.CacheItem{
		Key:        key,
		Value:      value,
		Priority:   priority,
		ExpireTime: expiry,
	}

	// push new item into priority queue for expiry
	heap.Push(&pe.PQueueForExpiry, &model.PriorityQueueItem{Item: newItem})
	// push the same new item into priority queue
	heap.Push(&pe.PriorityQueue, &model.PriorityQueueItem{Item: newItem})

	// push new item into the linked list
	element := pe.LRUList.PushFront(newItem)
	// add the item to map
	pe.ItemMap[key] = element

	// if cache is at max capacity evict an item
	if len(pe.ItemMap) > pe.MaxItems {
		pe.evictItem()
	}
}

// Get get item if exists and not expired
func (pe *PriorityExpiryCache) Get(key string) (interface{}, bool) {
	if element, found := pe.ItemMap[key]; found {
		pe.LRUList.MoveToFront(element) // move to the front of linked list
		item := element.Value.(*model.CacheItem)
		if time.Now().After(item.ExpireTime) {
			pe.removeElement(element, -1, -1)
			return nil, false
		}
		return item.Value, true
	}
	return nil, false
}

// evictItem will evict an item from the cache to make room for a new item.
func (pe *PriorityExpiryCache) evictItem() {
	if len(pe.ItemMap) <= pe.MaxItems {
		return
	}

	// delete the first item in the priority queue if it has expired
	if pe.PQueueForExpiry.Len() > 0 {
		priorityItem := pe.PQueueForExpiry[0] // peek at the first item
		if time.Now().After(priorityItem.Item.ExpireTime) {
			pe.remove(priorityItem.Item.Key, priorityItem.Index, -1)
			return
		}
	}

	// if no items has expired, get the item with the lowest priority and delete it
	for pe.PriorityQueue.Len() > 0 {
		// peek at the first and second item to check if we have more than one item with the same priority
		priorityItem1 := pe.PriorityQueue[0]
		priorityItem2 := pe.PriorityQueue[1]
		if priorityItem1.Item.Priority == priorityItem2.Item.Priority {
			break // since we have more than one we have to delete based on least recently used item
		}
		// else remove lowest priority item
		pe.remove(priorityItem1.Item.Key, -1, priorityItem1.Index)
		return
	}

	// delete least recently used item
	element := pe.LRUList.Back()
	pe.removeElement(element, -1, -1)
}

func (pe *PriorityExpiryCache) remove(key string, pqExpiryIndex, pqIndex int) {
	if element, found := pe.ItemMap[key]; found {
		pe.removeElement(element, pqExpiryIndex, pqIndex)
	}
}

func (pe *PriorityExpiryCache) removeElement(element *list.Element, pqExpiryIndex, pqIndex int) {
	item := element.Value.(*model.CacheItem)
	delete(pe.ItemMap, item.Key)
	pe.LRUList.Remove(element)

	// if we have the index use it to remove item from the priority queue
	if pqExpiryIndex > -1 {
		heap.Remove(&pe.PQueueForExpiry, pqExpiryIndex)
	} else {
		for i, pqeItem := range pe.PQueueForExpiry {
			if pqeItem.Item.Key == item.Key {
				heap.Remove(&pe.PQueueForExpiry, i)
				break
			}
		}
	}

	// if we have the index use it to remove item from the priority queue
	if pqIndex > -1 {
		heap.Remove(&pe.PriorityQueue, pqIndex)
	} else {
		for i, pqItem := range pe.PriorityQueue {
			if pqItem.Item.Key == item.Key {
				heap.Remove(&pe.PriorityQueue, i)
				break
			}
		}
	}
}

// update both the priority queues
func (pe *PriorityExpiryCache) update(key string) {
	if element, found := pe.ItemMap[key]; found {
		item := element.Value.(*model.CacheItem)

		for i, pqeItem := range pe.PQueueForExpiry {
			if pqeItem.Item.Key == key {
				pqeItem.Item.ExpireTime = item.ExpireTime
				heap.Fix(&pe.PQueueForExpiry, i)
				break
			}
		}

		for i, pqItem := range pe.PriorityQueue {
			if pqItem.Item.Key == key {
				pqItem.Item.Priority = item.Priority
				heap.Fix(&pe.PriorityQueue, i)
				break
			}
		}
	}
}

func main() {
	cache := NewPriorityExpiryCache(3)

	/* Test 1 - remove item that has expired
	1. Max Capacity of the cache is 3.
	2. Add 3 items A,B,C with same priority
	3. Add item D, since cache is at capacity (3) item B gets removed since it has expired.
	*/
	fmt.Println("Test 1")
	cache.Set("A", "Item A", 1, time.Second*5)
	cache.Set("B", "Item B", 1, time.Second*2)
	cache.Set("C", "Item C", 1, time.Second*5)
	time.Sleep(time.Second * 2)
	// add D
	cache.Set("D", "Item D", 3, time.Second*5)

	// expect Key B not found message since B was evicted
	if val, found := cache.Get("B"); found {
		fmt.Printf("\nItem B found, value: %s", val)
	} else {
		fmt.Println("\nItem B not found")
	}

	/* Test 2 - remove Item with the least priority
	1. Max Capacity of the cache is 3.
	2. Update items A,C,D. C has the lowest priority.
	3. Add item B, since cache is at capacity (3) item C gets removed since C has the lowest priority
	*/
	fmt.Println("\nTest 2")
	// the items get updated using the update method.
	cache.Set("A", "Item A", 3, time.Second*5)
	cache.Set("C", "Item C", 1, time.Second*6) // lowest priority
	cache.Set("D", "Item D", 3, time.Second*5)
	// add B
	cache.Set("B", "Item B", 3, time.Second*5)

	// expect Key A found message
	if val, found := cache.Get("A"); found {
		fmt.Printf("\nKey A found, value: %s", val)
	} else {
		fmt.Println("\nKey A not found")
	}

	// expect Key C not found message
	if val, found := cache.Get("C"); found {
		fmt.Printf("\nKey C found, value: %s", val)
	} else {
		fmt.Println("\nKey C not found")
	}

	/* Test 3 - remove least recently used item
	1. Max Capacity of the cache is 3.
	2. Update items A,B,D. B and D have the same priority.
	3. Get B
	4. Reduce cache capacity to 2. Since item A was least recently used, it gets evicted
	*/
	fmt.Println("\nTest 3")
	// the items get updated using the update method.
	cache.Set("A", "Item A", 3, time.Second*30)
	cache.Set("B", "Item B", 1, time.Second*20)
	cache.Set("D", "Item D", 1, time.Second*20)

	// expect Key B found message
	if val, found := cache.Get("B"); found {
		fmt.Printf("\nKey B found, value: %s", val)
	} else {
		fmt.Println("\nKey B not found")
	}

	// reduce capacity to 2
	cache.SetMaxItems(2)

	// expect Key A not found message
	if val, found := cache.Get("A"); found {
		fmt.Printf("\nKey A found, value: %s", val)
	} else {
		fmt.Println("\nKey A not found")
	}
}
