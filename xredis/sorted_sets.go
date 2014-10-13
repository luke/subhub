package xredis

// ZAdd adds all the specified members with the specified scores to the sorted set stored at key.
// If a specified member is already a member of the sorted set,
// the score is updated and the element reinserted at the right position to ensure the correct ordering.
// If key does not exist, a new sorted set with the specified members as sole members is created,
// like if the sorted set was empty.
// If the key exists but does not hold a sorted set, an error is returned.
//
// Return value:
// The number of elements added to the sorted sets,
// not including elements already existing for which the score was updated.
func (r *Redis) ZAdd(key string, pairs map[string]float64) (int64, error) {
	return r.redisMaster.ZAdd(key, pairs)
}

// ZCard returns the sorted set cardinality (number of elements) of the sorted set stored at key.
// Integer reply: the cardinality (number of elements) of the sorted set, or 0 if key does not exist.
func (r *Redis) ZCard(key string) (int64, error) {
	return r.redisSlave.ZCard(key)
}

// ZCount returns the number of elements in the sorted set at key with a score between min and max.
// The min and max arguments have the same semantic as described for ZRANGEBYSCORE.
// Integer reply: the number of elements in the specified score range.
func (r *Redis) ZCount(key, min, max string) (int64, error) {
	return r.redisSlave.ZCount(key, min, max)
}

// ZIncrBy increments the score of member in the sorted set stored at key by increment.
// If member does not exist in the sorted set, it is added with increment as its score
// (as if its previous score was 0.0).
// If key does not exist, a new sorted set with the specified member as its sole member is created.
// An error is returned when key exists but does not hold a sorted set.
// Bulk reply: the new score of member (a double precision floating point number), represented as string.
func (r *Redis) ZIncrBy(key string, increment float64, member string) (float64, error) {
	return r.redisMaster.ZIncrBy(key, increment, member)
}

// ZInterStore destination numkeys key [key ...] [WEIGHTS weight [weight ...]] [AGGREGATE SUM|MIN|MAX]
func (r *Redis) ZInterStore(destination string, keys []string, weights []int, aggregate string) (int64, error) {
	return r.redisMaster.ZInterStore(destination, keys, weights, aggregate)
}

// ZLexCount returns the number of elements in the sorted set at key
// with a value between min and max in order to force lexicographical ordering.
func (r *Redis) ZLexCount(key, min, max string) (int64, error) {
	return r.redisSlave.ZLexCount(key, min, max)
}

// ZRange returns the specified range of elements in the sorted set stored at key.
// The elements are considered to be ordered from the lowest to the highest score.
// Lexicographical order is used for elements with equal score.
// Multi-bulk reply: list of elements in the specified range.(optionally with their scores).
// It is possible to pass the WITHSCORES option in order to return the scores of the elements
// together with the elements.
// The returned list will contain value1,score1,...,valueN,scoreN instead of value1,...,valueN.
func (r *Redis) ZRange(key string, start, stop int, withscores bool) ([]string, error) {
	return r.redisSlave.ZRange(key, start, stop, withscores)
}

// ZRangeByLex returns all the elements in the sorted set at key with a value between min and max
// in order to force lexicographical ordering.
func (r *Redis) ZRangeByLex(key, min, max string, limit bool, offset, count int) ([]string, error) {
	return r.redisSlave.ZRangeByLex(key, min, max, limit, offset, count)
}

// ZRangeByScore key min max [WITHSCORES] [LIMIT offset count]
func (r *Redis) ZRangeByScore(key, min, max string, withscores, limit bool, offset, count int) ([]string, error) {
	return r.redisSlave.ZRangeByScore(key, min, max, withscores, limit, offset, count)
}

// ZRank returns the rank of member in the sorted set stored at key,
// with the scores ordered from low to high.
// The rank (or index) is 0-based, which means that the member with the lowest score has rank 0.
//
// If member exists in the sorted set, Integer reply: the rank of member.
// If member does not exist in the sorted set or key does not exist, Bulk reply: nil.
// -1 represent the nil bulk rely.
func (r *Redis) ZRank(key, member string) (int64, error) {
	return r.redisSlave.ZRank(key, member)
}

// ZRem removes the specified members from the sorted set stored at key. Non existing members are ignored.
// An error is returned when key exists and does not hold a sorted set.
// Integer reply, specifically:
// The number of members removed from the sorted set, not including non existing members.
func (r *Redis) ZRem(key string, members ...string) (int64, error) {
	return r.redisMaster.ZRem(key, members...)
}

// ZRemRangeByLex removes all elements in the sorted set stored at key
// between the lexicographical range specified by min and max.
func (r *Redis) ZRemRangeByLex(key, min, max string) (int64, error) {
	return r.redisMaster.ZRemRangeByLex(key, min, max)
}

// ZRemRangeByRank removes all elements in the sorted set stored at key with rank between start and stop.
// Both start and stop are 0 -based indexes with 0 being the element with the lowest score.
// These indexes can be negative numbers, where they indicate offsets starting at the element with the highest score.
// For example: -1 is the element with the highest score, -2 the element with the second highest score and so forth.
// Integer reply: the number of elements removed.
func (r *Redis) ZRemRangeByRank(key string, start, stop int) (int64, error) {
	return r.redisMaster.ZRemRangeByRank(key, start, stop)
}

// ZRemRangeByScore removes all elements in the sorted set stored at key with a score between min and max (inclusive).
// Integer reply: the number of elements removed.
func (r *Redis) ZRemRangeByScore(key, min, max string) (int64, error) {
	return r.redisMaster.ZRemRangeByScore(key, min, max)
}

// ZRevRange returns the specified range of elements in the sorted set stored at key.
// The elements are considered to be ordered from the highest to the lowest score.
// Descending lexicographical order is used for elements with equal score.
// Multi-bulk reply: list of elements in the specified range (optionally with their scores).
func (r *Redis) ZRevRange(key string, start, stop int, withscores bool) ([]string, error) {
	return r.redisSlave.ZRevRange(key, start, stop, withscores)
}

// ZRevRangeByScore key max min [WITHSCORES] [LIMIT offset count]
func (r *Redis) ZRevRangeByScore(key, max, min string, withscores, limit bool, offset, count int) ([]string, error) {
	return r.redisSlave.ZRevRangeByScore(key, max, min, withscores, limit, offset, count)
}

// ZRevRank returns the rank of member in the sorted set stored at key,
// with the scores ordered from high to low. The rank (or index) is 0-based,
// which means that the member with the highest score has rank 0.
func (r *Redis) ZRevRank(key, member string) (int64, error) {
	return r.redisSlave.ZRevRank(key, member)
}

// ZScore returns the score of member in the sorted set at key.
// If member does not exist in the sorted set, or key does not exist, nil is returned.
// Bulk reply: the score of member (a double precision floating point number), represented as string.
func (r *Redis) ZScore(key, member string) ([]byte, error) {
	return r.redisSlave.ZScore(key, member)
}

// ZUnionStore destination numkeys key [key ...] [WEIGHTS weight [weight ...]] [AGGREGATE SUM|MIN|MAX]
func (r *Redis) ZUnionStore(destination string, keys []string, weights []int, aggregate string) (int64, error) {
	return r.redisMaster.ZUnionStore(destination, keys, weights, aggregate)
}

// ZScan key cursor [MATCH pattern] [COUNT count]
func (r *Redis) ZScan(key string, cursor uint64, pattern string, count int) (uint64, []string, error) {
	return r.redisSlave.ZScan(key, cursor, pattern, count)
}
