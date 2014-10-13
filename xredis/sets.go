package xredis

// SAdd add the specified members to the set stored at key.
// Specified members that are already a member of this set are ignored.
// If key does not exist, a new set is created before adding the specified members.
// An error is returned when the value stored at key is not a set.
//
// Integer reply: the number of elements that were added to the set,
// not including all the elements already present into the set.
func (r *Redis) SAdd(key string, members ...string) (int64, error) {
	return r.redisMaster.SAdd(key, members...)
}

// SCard returns the set cardinality (number of elements) of the set stored at key.
func (r *Redis) SCard(key string) (int64, error) {
	return r.redisSlave.SCard(key)
}

// SDiff returns the members of the set resulting from the difference
// between the first set and all the successive sets.
// Keys that do not exist are considered to be empty sets.
// Multi-bulk reply: list with members of the resulting set.
func (r *Redis) SDiff(keys ...string) ([]string, error) {
	return r.redisSlave.SDiff(keys...)
}

// SDiffStore is equal to SDIFF, but instead of returning the resulting set,
// it is stored in destination.
// If destination already exists, it is overwritten.
// Integer reply: the number of elements in the resulting set.
func (r *Redis) SDiffStore(destination string, keys ...string) (int64, error) {
	return r.redisMaster.SDiffStore(destination, keys...)
}

// SInter returns the members of the set resulting from the intersection of all the given sets.
// Multi-bulk reply: list with members of the resulting set.
func (r *Redis) SInter(keys ...string) ([]string, error) {
	return r.redisSlave.SInter(keys...)
}

// SInterStore is equal to SINTER, but instead of returning the resulting set,
// it is stored in destination.
// If destination already exists, it is overwritten.
// Integer reply: the number of elements in the resulting set.
func (r *Redis) SInterStore(destination string, keys ...string) (int64, error) {
	return r.redisMaster.SInterStore(destination, keys...)
}

// SIsMember returns if member is a member of the set stored at key.
func (r *Redis) SIsMember(key, member string) (bool, error) {
	return r.redisSlave.SIsMember(key, member)
}

// SMembers returns all the members of the set value stored at key.
func (r *Redis) SMembers(key string) ([]string, error) {
	return r.redisSlave.SMembers(key)
}

// SMove moves member from the set at source to the set at destination.
// This operation is atomic.
// In every given moment the element will appear to be a member of source or destination for other clients.
func (r *Redis) SMove(source, destination, member string) (bool, error) {
	return r.redisMaster.SMove(source, destination, member)
}

// SPop removes and returns a random element from the set value stored at key.
// Bulk reply: the removed element, or nil when key does not exist.
func (r *Redis) SPop(key string) ([]byte, error) {
	return r.redisMaster.SPop(key)
}

// SRandMember returns a random element from the set value stored at key.
// Bulk reply: the command returns a Bulk Reply with the randomly selected element,
// or nil when key does not exist.
func (r *Redis) SRandMember(key string) ([]byte, error) {
	return r.redisSlave.SRandMember(key)
}

// SRandMemberCount returns an array of count distinct elements if count is positive.
// If called with a negative count the behavior changes and the command
// is allowed to return the same element multiple times.
// In this case the numer of returned elements is the absolute value of the specified count.
// returns an array of elements, or an empty array when key does not exist.
func (r *Redis) SRandMemberCount(key string, count int) ([]string, error) {
	return r.redisSlave.SRandMemberCount(key, count)
}

// SRem remove the specified members from the set stored at key.
// Specified members that are not a member of this set are ignored.
// If key does not exist, it is treated as an empty set and this command returns 0.
// An error is returned when the value stored at key is not a set.
// Integer reply: the number of members that were removed from the set,
// not including non existing members.
func (r *Redis) SRem(key string, members ...string) (int64, error) {
	return r.redisMaster.SRem(key, members...)
}

// SUnion returns the members of the set resulting from the union of all the given sets.
// Multi-bulk reply: list with members of the resulting set.
func (r *Redis) SUnion(keys ...string) ([]string, error) {
	return r.redisSlave.SUnion(keys...)
}

// SUnionStore is equal to SUnion.
// If destination already exists, it is overwritten.
// Integer reply: the number of elements in the resulting set.
func (r *Redis) SUnionStore(destination string, keys ...string) (int64, error) {
	return r.redisMaster.SUnionStore(destination, keys...)
}

// SScan key cursor [MATCH pattern] [COUNT count]
func (r *Redis) SScan(key string, cursor uint64, pattern string, count int) (uint64, []string, error) {
	return r.redisSlave.SScan(key, cursor, pattern, count)
}
