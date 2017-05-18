package datatypes

type MapFn func(key string, value string, emitter Emitter)
type RedFn func(key string, values []string, emitter Emitter)

//The user must supply a Map function and a Reduce function.
//The default distributor for the map function is a hash-based distributor,
//and the default for the reduce function is a round robin.
type Job struct {
	Map           MapFn
	Reduce        RedFn
	MapDistribute Distributor
	RedDistribute Distributor
}
