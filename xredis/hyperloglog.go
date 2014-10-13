package xredis

// PFAdd adds all the element arguments to the HyperLogLog data structure
// stored at the variable name specified as first argument.
func (r *Redis) PFAdd(key string, elements ...string) (int64, error) {
	return r.redisMaster.PFAdd(key, elements...)
}

// PFCount returns the approximated cardinality computed by the HyperLogLog
// data structure stored at the specified variable,
// which is 0 if the variable does not exist.
// When called with multiple keys, returns the approximated cardinality of
// the union of the HyperLogLogs passed, by internally merging the HyperLogLogs
// stored at the provided keys into a temporary hyperLogLog.
func (r *Redis) PFCount(keys ...string) (int64, error) {
	return r.redisSlave.PFCount(keys...)
}

// PFMerge merges multiple HyperLogLog values into an unique value
// that will approximate the cardinality of the union of the observed
// Sets of the source HyperLogLog structures.
// The computed merged HyperLogLog is set to the destination variable,
// which is created if does not exist (defauling to an empty HyperLogLog).
func (r *Redis) PFMerge(destkey string, sourcekeys ...string) error {
	return r.redisMaster.PFMerge(destkey, sourcekeys...)
}
