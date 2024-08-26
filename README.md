# Priority Expiry Cache 
To implement this I have used the following datastructures.
1. Map: to Get and Set the items. And manage the capacity
2. Doubly Linked List: to get the least recently used item
3. Two Min Heap - Priority Queues: To get the item with the least priority and the expired item.
4. The model.go file has the structs and heap declaration.

Time Complexity - O(logN) for most operations, including adding, editing, evicting, removing and updating items.
Space Complexity - O(N) for storing up to N items in the cache.
