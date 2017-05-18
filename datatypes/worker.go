package datatypes

//A worker is a single goroutine running a single map or reduce function.
//It contains an input channel to listen on and a list of all of the channels
//after it to send its data to.
type worker interface {
	run()
	init(numUpstream int, inChannel chan [2]string, endpoints []chan [2]string)
}

type mapWorker struct {
	numUpstream int
	inChannel   chan [2]string
	endpoints   []chan [2]string
	distribute  Distributor
	Map         MapFn
}

func (mw *mapWorker) Emit(key string, value string) {
	mw.distribute([2]string{key, value}, mw.endpoints)
}
func (mw *mapWorker) run() {
	for {
		data := <-mw.inChannel
		if data[0] == "\x00" {
			mw.numUpstream--
			if mw.numUpstream == 0 {
				break
			}
			continue
		}
		mw.Map(data[0], data[1], mw)
	}
	end(mw.endpoints)
}

func (mw *mapWorker) init(numUpstream int, inChannel chan [2]string, endpoints []chan [2]string) {
	mw.numUpstream = numUpstream
	mw.inChannel = inChannel
	mw.endpoints = endpoints
}

type redWorker struct {
	numUpstream int
	inChannel   chan [2]string
	endpoints   []chan [2]string
	distribute  Distributor
	Reduce      RedFn

	buffers map[string][]string
}

func (rw *redWorker) Emit(key string, value string) {
	rw.distribute([2]string{key, value}, rw.endpoints)
}
func (rw *redWorker) run() {
	for {
		data := <-rw.inChannel
		if data[0] == "\x00" {
			rw.numUpstream--
			if rw.numUpstream == 0 {
				break
			}
			continue
		}
		rw.buffers[data[0]] = append(rw.buffers[data[0]], data[1])
	}
	for key, values := range rw.buffers {
		rw.Reduce(key, values, rw)
	}
	end(rw.endpoints)
}

func (rw *redWorker) init(numUpstream int, inChannel chan [2]string, endpoints []chan [2]string) {
	rw.numUpstream = numUpstream
	rw.inChannel = inChannel
	rw.endpoints = endpoints

	rw.buffers = make(map[string][]string)
}
