package xredis

// BLPop is a blocking list pop primitive.
// It is the blocking version of LPOP
// because it blocks the connection when there are no elements to pop from any of the given lists.
// An element is popped from the head of the first list that is non-empty,
// with the given keys being checked in the order that they are given.
// A nil multi-bulk when no element could be popped and the timeout expired.
// A two-element multi-bulk with the first element being the name of the key where an element was popped
// and the second element being the value of the popped element.
func (r *Redis) BLPop(keys []string, timeout int) ([]string, error) {
	return r.redisMaster.BLPop(keys, timeout)
}

// BRPop pops elements from the tail of a list instead of popping from the head.
func (r *Redis) BRPop(keys []string, timeout int) ([]string, error) {
	return r.redisMaster.BRPop(keys, timeout)
}

// BRPopLPush is the blocking variant of RPOPLPUSH.
// When source contains elements,
// this command behaves exactly like RPOPLPUSH.
// When source is empty, Redis will block the connection until
// another client pushes to it or until timeout is reached.
// A timeout of zero can be used to block indefinitely.
// Bulk reply: the element being popped from source and pushed to destination.
// If timeout is reached, a Null multi-bulk reply is returned.
func (r *Redis) BRPopLPush(source, destination string, timeout int) ([]byte, error) {
	return r.redisMaster.BRPopLPush(source, destination, timeout)
}

// LIndex returns the element at index index in the list stored at key.
// The index is zero-based, so 0 means the first element,
// 1 the second element and so on.
// Negative indices can be used to designate elements starting at the tail of the list.
// Here, -1 means the last element, -2 means the penultimate and so forth.
// When the value at key is not a list, an error is returned.
// Bulk reply: the requested element, or nil when index is out of range.
func (r *Redis) LIndex(key string, index int) ([]byte, error) {
	return r.redisSlave.LIndex(key, index)
}

// LInsert inserts value in the list stored at key either before or after the reference value pivot.
// When key does not exist, it is considered an empty list and no operation is performed.
// An error is returned when key exists but does not hold a list value.
// Integer reply: the length of the list after the insert operation, or -1 when the value pivot was not found.
func (r *Redis) LInsert(key, position, pivot, value string) (int64, error) {
	return r.redisMaster.LInsert(key, position, pivot, value)
}

// LLen returns the length of the list stored at key.
// If key does not exist, it is interpreted as an empty list and 0 is returned.
// An error is returned when the value stored at key is not a list.
func (r *Redis) LLen(key string) (int64, error) {
	return r.redisSlave.LLen(key)
}

// LPop removes and returns the first element of the list stored at key.
// Bulk reply: the value of the first element, or nil when key does not exist.
func (r *Redis) LPop(key string) ([]byte, error) {
	return r.redisMaster.LPop(key)
}

// LPush insert all the specified values at the head of the list stored at key.
// If key does not exist, it is created as empty list before performing the push operations.
// When key holds a value that is not a list, an error is returned.
// Integer reply: the length of the list after the push operations.
func (r *Redis) LPush(key string, values ...string) (int64, error) {
	return r.redisMaster.LPush(key, values...)
}

// LPushx inserts value at the head of the list stored at key,
// only if key already exists and holds a list.
// In contrary to LPUSH, no operation will be performed when key does not yet exist.
// Integer reply: the length of the list after the push operation.
func (r *Redis) LPushx(key, value string) (int64, error) {
	return r.redisMaster.LPushx(key, value)
}

// LRange returns the specified elements of the list stored at key.
// The offsets start and stop are zero-based indexes,
// with 0 being the first element of the list (the head of the list), 1 being the next element and so on.
// These offsets can also be negative numbers indicating offsets starting at the end of the list.
// For example, -1 is the last element of the list, -2 the penultimate, and so on.
//
// Note that if you have a list of numbers from 0 to 100, LRANGE list 0 10 will return 11 elements,
// that is, the rightmost item is included.
// Out of range indexes will not produce an error.
// If start is larger than the end of the list, an empty list is returned.
// If stop is larger than the actual end of the list, Redis will treat it like the last element of the list.
// Multi-bulk reply: list of elements in the specified range.
func (r *Redis) LRange(key string, start, end int) ([]string, error) {
	return r.redisSlave.LRange(key, start, end)
}

// LRem removes the first count occurrences of elements equal to value from the list stored at key.
// The count argument influences the operation in the following ways:
// count > 0: Remove elements equal to value moving from head to tail.
// count < 0: Remove elements equal to value moving from tail to head.
// count = 0: Remove all elements equal to value.
// Integer reply: the number of removed elements.
func (r *Redis) LRem(key string, count int, value string) (int64, error) {
	return r.redisMaster.LRem(key, count, value)
}

// LSet sets the list element at index to value. For more information on the index argument, see LINDEX.
// An error is returned for out of range indexes.
func (r *Redis) LSet(key string, index int, value string) error {
	return r.redisMaster.LSet(key, index, value)
}

// LTrim trim an existing list so that it will contain only the specified range of elements specified.
// Both start and stop are zero-based indexes, where 0 is the first element of the list (the head),
// 1 the next element and so on.
func (r *Redis) LTrim(key string, start, stop int) error {
	return r.redisMaster.LTrim(key, start, stop)
}

// RPop removes and returns the last element of the list stored at key.
// Bulk reply: the value of the last element, or nil when key does not exist.
func (r *Redis) RPop(key string) ([]byte, error) {
	return r.redisMaster.RPop(key)
}

// RPopLPush atomically returns and removes the last element (tail) of the list stored at source,
// and pushes the element at the first element (head) of the list stored at destination.
//
// If source does not exist, the value nil is returned and no operation is performed.
// If source and destination are the same,
// the operation is equivalent to removing the last element from the list and pushing it as first element of the list,
// so it can be considered as a list rotation command.
func (r *Redis) RPopLPush(source, destination string) ([]byte, error) {
	return r.redisMaster.RPopLPush(source, destination)
}

// RPush insert all the specified values at the tail of the list stored at key.
// If key does not exist, it is created as empty list before performing the push operation.
// When key holds a value that is not a list, an error is returned.
func (r *Redis) RPush(key string, values ...string) (int64, error) {
	return r.redisMaster.RPush(key, values...)
}

// RPushx inserts value at the tail of the list stored at key,
// only if key already exists and holds a list.
// In contrary to RPUSH, no operation will be performed when key does not yet exist.
func (r *Redis) RPushx(key, value string) (int64, error) {
	return r.redisMaster.RPushx(key, value)
}
