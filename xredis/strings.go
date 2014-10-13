package xredis

// Append appends the value at the end of the string which stored at key
// If key does not exist it is created and set as an empty string.
// Return integer reply: the length of the string after the append operation.
func (r *Redis) Append(key, value string) (int64, error) {
	return r.redisMaster.Append(key, value)
}

// BitCount counts the number of set bits (population counting) in a string.
func (r *Redis) BitCount(key string, start, end int) (int64, error) {
	return r.redisSlave.BitCount(key, start, end)
}

// BitOp performs a bitwise operation between multiple keys (containing string values)
// and store the result in the destination key.
// The BITOP command supports four bitwise operations:
// AND, OR, XOR and NOT, thus the valid forms to call the command are:
// BITOP AND destkey srckey1 srckey2 srckey3 ... srckeyN
// BITOP OR destkey srckey1 srckey2 srckey3 ... srckeyN
// BITOP XOR destkey srckey1 srckey2 srckey3 ... srckeyN
// BITOP NOT destkey srckey
// Return value: Integer reply
// The size of the string stored in the destination key, that is equal to the size of the longest input string.
func (r *Redis) BitOp(operation, destkey string, keys ...string) (int64, error) {
	return r.redisMaster.BitOp(operation, destkey, keys...)
}

// Decr decrements the number stored at key by one.
// If the key does not exist, it is set to 0 before performing the operation.
// An error is returned if the key contains a value of the wrong type
// or contains a string that can not be represented as integer.
// This operation is limited to 64 bit signed integers.
// Integer reply: the value of key after the decrement
func (r *Redis) Decr(key string) (int64, error) {
	return r.redisMaster.Decr(key)
}

// DecrBy decrements the number stored at key by decrement.
func (r *Redis) DecrBy(key string, decrement int) (int64, error) {
	return r.redisMaster.DecrBy(key, decrement)
}

// Get gets the value of key.
// If the key does not exist the special value nil is returned.
// An error is returned if the value stored at key is not a string,
// because GET only handles string values.
func (r *Redis) Get(key string) ([]byte, error) {
	return r.redisSlave.Get(key)
}

// GetBit returns the bit value at offset in the string value stored at key.
// When offset is beyond the string length,
// the string is assumed to be a contiguous space with 0 bits.
// When key does not exist it is assumed to be an empty string,
// so offset is always out of range and the value is also assumed to be a contiguous space with 0 bits.
func (r *Redis) GetBit(key string, offset int) (int64, error) {
	return r.redisSlave.GetBit(key, offset)
}

// GetRange returns the substring of the string value stored at key,
// determined by the offsets start and end (both are inclusive).
// Negative offsets can be used in order to provide an offset starting from the end of the string.
// So -1 means the last character, -2 the penultimate and so forth.
// The function handles out of range requests by limiting the resulting range to the actual length of the string.
func (r *Redis) GetRange(key string, start, end int) (string, error) {
	return r.redisSlave.GetRange(key, start, end)
}

// GetSet atomically sets key to value and returns the old value stored at key.
// Returns an error when key exists but does not hold a string value.
func (r *Redis) GetSet(key, value string) ([]byte, error) {
	return r.redisMaster.GetSet(key, value)
}

// Incr increments the number stored at key by one.
// If the key does not exist, it is set to 0 before performing the operation.
// An error is returned if the key contains a value of the wrong type
// or contains a string that can not be represented as integer.
// Integer reply: the value of key after the increment
func (r *Redis) Incr(key string) (int64, error) {
	return r.redisMaster.Incr(key)
}

// IncrBy increments the number stored at key by increment.
// If the key does not exist, it is set to 0 before performing the operation.
// An error is returned if the key contains a value of the wrong type
// or contains a string that can not be represented as integer.
// Integer reply: the value of key after the increment
func (r *Redis) IncrBy(key string, increment int) (int64, error) {
	return r.redisMaster.IncrBy(key, increment)
}

// IncrByFloat increments the string representing a floating point number
// stored at key by the specified increment.
// If the key does not exist, it is set to 0 before performing the operation.
// An error is returned if one of the following conditions occur:
// The key contains a value of the wrong type (not a string).
// The current key content or the specified increment are not parsable
// as a double precision floating point number.
// Return bulk reply: the value of key after the increment.
func (r *Redis) IncrByFloat(key string, increment float64) (float64, error) {
	return r.redisMaster.IncrByFloat(key, increment)
}

// MGet returns the values of all specified keys.
// For every key that does not hold a string value or does not exist,
// the special value nil is returned. Because of this, the operation never fails.
// Multi-bulk reply: list of values at the specified keys.
func (r *Redis) MGet(keys ...string) ([][]byte, error) {
	return r.redisSlave.MGet(keys...)
}

// MSet sets the given keys to their respective values.
// MSET replaces existing values with new values, just as regular SET.
// See MSETNX if you don't want to overwrite existing values.
func (r *Redis) MSet(pairs map[string]string) error {
	return r.redisMaster.MSet(pairs)
}

// MSetnx sets the given keys to their respective values.
// MSETNX will not perform any operation at all even if just a single key already exists.
// True if the all the keys were set.
// False if no key was set (at least one key already existed).
func (r *Redis) MSetnx(pairs map[string]string) (bool, error) {
	return r.redisMaster.MSetnx(pairs)
}

// PSetex works exactly like SETEX with the sole difference that
// the expire time is specified in milliseconds instead of seconds.
func (r *Redis) PSetex(key string, milliseconds int, value string) error {
	return r.redisMaster.PSetex(key, milliseconds, value)
}

// Set sets key to hold the string value.
// If key already holds a value, it is overwritten, regardless of its type.
// Any previous time to live associated with the key is discarded on successful SET operation.
func (r *Redis) Set(key, value string, seconds, milliseconds int, mustExists, mustNotExists bool) error {
	return r.redisMaster.Set(key, value, seconds, milliseconds, mustExists, mustNotExists)
}

// SimpleSet do SET key value, no other arguments.
func (r *Redis) SimpleSet(key, value string) error {
	return r.redisMaster.SimpleSet(key, value)
}

// SetBit sets or clears the bit at offset in the string value stored at key.
// Integer reply: the original bit value stored at offset.
func (r *Redis) SetBit(key string, offset, value int) (int64, error) {
	return r.redisMaster.SetBit(key, offset, value)
}

// Setex sets key to hold the string value and set key to timeout after a given number of seconds.
func (r *Redis) Setex(key string, seconds int, value string) error {
	return r.redisMaster.Setex(key, seconds, value)
}

// Setnx sets key to hold string value if key does not exist.
func (r *Redis) Setnx(key, value string) (bool, error) {
	return r.redisMaster.Setnx(key, value)
}

// SetRange overwrites part of the string stored at key, starting at the specified offset,
// for the entire length of value.
// Integer reply: the length of the string after it was modified by the command.
func (r *Redis) SetRange(key string, offset int, value string) (int64, error) {
	return r.redisMaster.SetRange(key, offset, value)
}

// StrLen returns the length of the string value stored at key.
// An error is returned when key holds a non-string value.
// Integer reply: the length of the string at key, or 0 when key does not exist.
func (r *Redis) StrLen(key string) (int64, error) {
	return r.redisSlave.StrLen(key)
}
