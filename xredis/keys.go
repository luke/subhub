package xredis

import (
	"github.com/xuyu/goredis"
)

// Del removes the specified keys.
// A key is ignored if it does not exist.
// Integer reply: The number of keys that were removed.
func (r *Redis) Del(keys ...string) (int64, error) {
	return r.redisMaster.Del(keys...)
}

// Dump serialize the value stored at key in a Redis-specific format and return it to the user.
// The returned value can be synthesized back into a Redis key using the RESTORE command.
// Return []byte for maybe big data
func (r *Redis) Dump(key string) ([]byte, error) {
	return r.redisSlave.Dump(key)
}

// Exists returns true if key exists.
func (r *Redis) Exists(key string) (bool, error) {
	return r.redisSlave.Exists(key)
}

// Expire set a second timeout on key.
// After the timeout has expired, the key will automatically be deleted.
// A key with an associated timeout is often said to be volatile in Redis terminology.
func (r *Redis) Expire(key string, seconds int) (bool, error) {
	return r.redisMaster.Expire(key, seconds)
}

// ExpireAt has the same effect and semantic as expire,
// but instead of specifying the number of seconds representing the TTL (time to live),
// it takes an absolute Unix timestamp (seconds since January 1, 1970).
func (r *Redis) ExpireAt(key string, timestamp int64) (bool, error) {
	return r.redisMaster.ExpireAt(key, timestamp)
}

// Keys returns all keys matching pattern.
func (r *Redis) Keys(pattern string) ([]string, error) {
	return r.redisSlave.Keys(pattern)
}

// Atomically transfer a key from a source Redis instance to a destination Redis instance.
// On success the key is deleted from the original instance and is guaranteed to exist in the target instance.
//
// The command is atomic and blocks the two instances for the time required to transfer the key,
// at any given time the key will appear to exist in a given instance or in the other instance,
// unless a timeout error occurs.
//
// The timeout specifies the maximum idle time in any moment of the communication
// with the destination instance in milliseconds.
//
// COPY -- Do not remove the key from the local instance.
// REPLACE -- Replace existing key on the remote instance.
//
// Status code reply: The command returns OK on success.
// MIGRATE host port key destination-db timeout [COPY] [REPLACE]
//
// func (r *Redis) Migrate(host, port, key string, db, timeout int, cp, replace bool) error {
// 	args := packArgs("MIGRATE", host, port, key, db, timeout)
// 	if cp {
// 		args = append(args, "COPY")
// 	}
// 	if replace {
// 		args = append(args, "REPLACE")
// 	}
// 	rp, err := r.ExecuteCommand(args...)
// 	if err != nil {
// 		return err
// 	}
// 	return rp.OKValue()
// }

// Move moves key from the currently selected database (see SELECT)
// to the specified destination database.
// When key already exists in the destination database,
// or it does not exist in the source database, it does nothing.
func (r *Redis) Move(key string, db int) (bool, error) {
	return r.redisMaster.Move(key, db)
}

// Object inspects the internals of Redis Objects associated with keys.
// It is useful for debugging or to understand if your keys are using the specially encoded data types to save space.
// Your application may also use the information reported by the OBJECT command
// to implement application level key eviction policies
// when using Redis as a Cache.
func (r *Redis) Object(subcommand string, arguments ...string) (*goredis.Reply, error) {
	return r.redisSlave.Object(subcommand, arguments...)
}

// Persist removes the existing timeout on key,
// turning the key from volatile (a key with an expire set) to persistent
// (a key that will never expire as no timeout is associated).
// True if the timeout was removed.
// False if key does not exist or does not have an associated timeout.
func (r *Redis) Persist(key string) (bool, error) {
	return r.redisMaster.Persist(key)
}

// PExpire works exactly like EXPIRE
// but the time to live of the key is specified in milliseconds instead of seconds.
func (r *Redis) PExpire(key string, milliseconds int) (bool, error) {
	return r.redisMaster.PExpire(key, milliseconds)
}

// PExpireAt has the same effect and semantic as EXPIREAT,
// but the Unix time at which the key will expire is specified in milliseconds instead of seconds.
func (r *Redis) PExpireAt(key string, timestamp int64) (bool, error) {
	return r.redisMaster.PExpireAt(key, timestamp)
}

// PTTL returns the remaining time to live of a key that has an expire set,
// with the sole difference that TTL returns the amount of remaining time in seconds
// while PTTL returns it in milliseconds.
func (r *Redis) PTTL(key string) (int64, error) {
	return r.redisSlave.PTTL(key)
}

// RandomKey returns a random key from the currently selected database.
// Bulk reply: the random key, or nil when the database is empty.
func (r *Redis) RandomKey() ([]byte, error) {
	return r.redisSlave.RandomKey()
}

// Rename renames key to newkey.
// It returns an error when the source and destination names are the same, or when key does not exist.
// If newkey already exists it is overwritten, when this happens RENAME executes an implicit DEL operation,
// so if the deleted key contains a very big value it may cause high latency
// even if RENAME itself is usually a constant-time operation.
func (r *Redis) Rename(key, newkey string) error {
	return r.redisMaster.Rename(key, newkey)
}

// Renamenx renames key to newkey if newkey does not yet exist.
// It returns an error under the same conditions as RENAME.
func (r *Redis) Renamenx(key, newkey string) (bool, error) {
	return r.redisMaster.Renamenx(key, newkey)
}

// Restore creates a key associated with a value that is obtained by deserializing
// the provided serialized value (obtained via DUMP).
// If ttl is 0 the key is created without any expire, otherwise the specified expire time (in milliseconds) is set.
// RESTORE checks the RDB version and data checksum. If they don't match an error is returned.
func (r *Redis) Restore(key string, ttl int, serialized string) error {
	return r.redisMaster.Restore(key, ttl, serialized)
}

// TTL returns the remaining time to live of a key that has a timeout.
// Integer reply: TTL in seconds, or a negative value in order to signal an error (see the description above).
func (r *Redis) TTL(key string) (int64, error) {
	return r.redisSlave.TTL(key)
}

// Type returns the string representation of the type of the value stored at key.
// The different types that can be returned are: string, list, set, zset and hash.
// Status code reply: type of key, or none when key does not exist.
func (r *Redis) Type(key string) (string, error) {
	return r.redisSlave.Type(key)
}

// Scan command:
// SCAN cursor [MATCH pattern] [COUNT count]
func (r *Redis) Scan(cursor uint64, pattern string, count int) (uint64, []string, error) {
	return r.redisSlave.Scan(cursor, pattern, count)
}
