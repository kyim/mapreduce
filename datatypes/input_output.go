package datatypes

//Input is used to generate data to be processed.
type Input struct {
	endpoints []chan [2]string

	Param      string
	//GenInput is a single, user-defined function that emits all of the data
	//to be processed.
	GenInput   func(param string, emitter Emitter)
	Distribute Distributor
}

func (i *Input) Emit(key string, value string) {
	i.Distribute([2]string{key, value}, i.endpoints)
}

func (i *Input) run() {
	i.GenInput(i.Param, i)
	end(i.endpoints)
}

func (i *Input) init(endpoints []chan [2]string) {
	i.endpoints = endpoints
}

//Output is used to handle the data that has been processed.
type Output struct {
	numUpstream int
	inChannel   chan [2]string
	endChannel  chan int

	Param      string
	//InitOutput is run before the output starts accepting data.
	//The master will ensure this function is never nil.
	InitOutput func(param string)
	//GenOutput is a user-defined function to handle the data as it is generated.
	GenOutput  func(param, key, value string)
	//EndOutput is run when the output has finished accepting data.
	//The master will ensure this function is never nil.
	EndOutput  func()
}

func (o *Output) run() {
	o.InitOutput(o.Param)
	count := 0
	for {
		data := <-o.inChannel
		if data[0] == "\x00" {
			o.numUpstream--
			if o.numUpstream == 0 {
				break
			}
			continue
		}
		o.GenOutput(o.Param, data[0], data[1])
		count++
	}
	o.endChannel <- count
	o.EndOutput()
}

func (o *Output) init(numUpstream int, inChannel chan [2]string) {
	o.numUpstream = numUpstream
	o.inChannel = inChannel
	o.endChannel = make(chan int)
}
