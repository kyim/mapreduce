//Package datatypes provides the constructs for building and running
//mapreduce jobs. Most of the functionality is handled through the
//Master struct and methods.
package datatypes

func end(channels []chan [2]string) {
	for _, channel := range channels {
		channel <- [2]string{"\x00", ""}
	}
}

//Emitter is an interface for accepting key-value pairs from within a
//processing function, to be passed to the next function.
type Emitter interface {
	Emit(key string, value string)
}

//Distributors are functions that handle which channel receives a given
//key-value pair.
type Distributor func(data [2]string, channels []chan [2]string)
