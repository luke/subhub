package xredis

// HDel command:
// Removes the specified fields from the hash stored at key.
// Specified fields that do not exist within this hash are ignored.
// If key does not exist, it is treated as an empty hash and this command returns 0.
func (r *Redis) HDel(key string, fields ...string) (int64, error) {
	return r.redisMaster.HDel(key, fields...)
}

// HExists command:
// Returns if field is an existing field in the hash stored at key.
func (r *Redis) HExists(key, field string) (bool, error) {
	return r.redisSlave.HExists(key, field)
}

// HGet command:
// Returns the value associated with field in the hash stored at key.
// Bulk reply: the value associated with field,
// or nil when field is not present in the hash or key does not exist.
func (r *Redis) HGet(key, field string) ([]byte, error) {
	return r.redisSlave.HGet(key, field)
}

// HGetAll command:
// Returns all fields and values of the hash stored at key.
// In the returned value, every field name is followed by its value,
// so the length of the reply is twice the size of the hash.
func (r *Redis) HGetAll(key string) (map[string]string, error) {
	return r.redisSlave.HGetAll(key)
}

// HIncrBy command:
// Increments the number stored at field in the hash stored at key by increment.
// If key does not exist, a new key holding a hash is created.
// If field does not exist the value is set to 0 before the operation is performed.
// Integer reply: the value at field after the increment operation.
func (r *Redis) HIncrBy(key, field string, increment int) (int64, error) {
	return r.redisMaster.HIncrBy(key, field, increment)
}

// HIncrByFloat command:
// Increment the specified field of an hash stored at key,
// and representing a floating point number, by the specified increment.
// If the field does not exist, it is set to 0 before performing the operation.
// An error is returned if one of the following conditions occur:
// The field contains a value of the wrong type (not a string).
// The current field content or the specified increment are not parsable as a double precision floating point number.
// Bulk reply: the value of field after the increment.
func (r *Redis) HIncrByFloat(key, field string, increment float64) (float64, error) {
	return r.redisMaster.HIncrByFloat(key, field, increment)
}

// HKeys command:
// Returns all field names in the hash stored at key.
// Multi-bulk reply: list of fields in the hash, or an empty list when key does not exist.
func (r *Redis) HKeys(key string) ([]string, error) {
	return r.redisSlave.HKeys(key)
}

// HLen command:
// Returns the number of fields contained in the hash stored at key.
// Integer reply: number of fields in the hash, or 0 when key does not exist.
func (r *Redis) HLen(key string) (int64, error) {
	return r.redisSlave.HLen(key)
}

// HMGet command:
// Returns the values associated with the specified fields in the hash stored at key.
// For every field that does not exist in the hash, a nil value is returned.
// Because a non-existing keys are treated as empty hashes,
// running HMGET against a non-existing key will return a list of nil values.
// Multi-bulk reply: list of values associated with the given fields, in the same order as they are requested.
func (r *Redis) HMGet(key string, fields ...string) ([][]byte, error) {
	return r.redisSlave.HMGet(key, fields...)
}

// HMSet command:
// Sets the specified fields to their respective values in the hash stored at key.
// This command overwrites any existing fields in the hash.
// If key does not exist, a new key holding a hash is created.
func (r *Redis) HMSet(key string, pairs map[string]string) error {
	return r.redisMaster.HMSet(key, pairs)
}

// HSet command:
// Sets field in the hash stored at key to value.
// If key does not exist, a new key holding a hash is created.
// If field already exists in the hash, it is overwritten.
func (r *Redis) HSet(key, field, value string) (bool, error) {
	return r.redisMaster.HSet(key, field, value)
}

// HSetnx command:
// Sets field in the hash stored at key to value, only if field does not yet exist.
// If key does not exist, a new key holding a hash is created.
// If field already exists, this operation has no effect.
func (r *Redis) HSetnx(key, field, value string) (bool, error) {
	return r.redisMaster.HSetnx(key, field, value)
}

// HVals command:
// Returns all values in the hash stored at key.
// Multi-bulk reply: list of values in the hash, or an empty list when key does not exist.
func (r *Redis) HVals(key string) ([]string, error) {
	return r.redisSlave.HVals(key)
}

// HScan command:
// HSCAN key cursor [MATCH pattern] [COUNT count]
func (r *Redis) HScan(key string, cursor uint64, pattern string, count int) (uint64, map[string]string, error) {
	return r.redisSlave.HScan(key, cursor, pattern, count)
}
