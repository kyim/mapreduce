//Package inverted_index contains an example for building an inverted index on
//the received key-value pairs
package inverted_index

import (
	. "mapreduce/datatypes"
	"sort"
	"strings"
)

//The map function swaps the position of the key and value.
func MapIndex(key string, value string, emitter Emitter) {
	emitter.Emit(value, key) 
}

//The reduce function outputs the value (now the key) and all of the keys that
//were associated with that value
func ReduceIndex(key string, values []string, emitter Emitter) {
	sort.Strings(values)
	emitter.Emit("", key+": "+strings.Join(values, ", "))
}
