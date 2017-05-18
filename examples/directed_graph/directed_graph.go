//Package directed_graph contains an example for determining the cycles of length
//three in a directed graph. This requires two MapReduce iterations.
package directed_graph

import (
	. "mapreduce/datatypes"
	"strings"
)

//The first map iteration emits every edge twice, once for the start node and once
//for the end node.
func MapGraph1(key string, value string, emitter Emitter) {
	splits := strings.Split(value, ":")
	points := strings.Split(splits[1], ",")
	for _, point := range points {
		emitter.Emit(splits[0], splits[0]+","+point)
		emitter.Emit(point, splits[0]+","+point)
	}
}

//The first reduce iteration emits every path of length two, using the endpoints
//as the key (end first) and the midpoint as the value. It also emits the word
//"yes" for every path of length one, using the endpoints as the key (end second).
//There is a cycle of length three if and only if there is a path of length two
//from A to B and a path of length one from B to A. Obviously the choice of the
//word "yes" is arbitrary. The node names can be extended to accept any string
//(including "yes" and all possible flag values) if you implement some kind of
//unique prefix (i.e. put "1" before the name of the midpoint, and use "2yes"
//as the flag, or actually just "yes" because it does not start with "1")
func ReduceGraph1(key string, values []string, emitter Emitter) {
	var ins []string
	var outs []string
	for _, pair := range values {
		split := strings.Split(pair, ",")
		if split[0] == key {
			outs = append(outs, split[1])
		} else if split[1] == key {
			ins = append(ins, split[0])
		}

		emitter.Emit(pair, "yes")
	}
	for _, in := range ins {
		for _, out := range outs {
			emitter.Emit(out+","+in, key)
		}
	}
}

//The second map iteration is the identity function
func MapGraph2(key string, value string, emitter Emitter) {
	emitter.Emit(key, value)
}

//The second reduce iteration checks for cycles as described above, and outputs
//the if and only if the first node name is less than the others (to prevent
//repeating each cycle three times)
func ReduceGraph2(key string, values []string, emitter Emitter) {
	var output []string
	yes := false
	for _, value := range values {
		if value == "yes" {
			yes = true
		} else {
			split := strings.Split(key, ",")
			if value < split[0] && value < split[1] {
				output = append(output, value+","+split[0]+","+split[1])
			}
		}
	}
	if yes {
		for _, out := range output {
			emitter.Emit("", out)
		}
	}
}
