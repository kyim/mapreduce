package datatypes

//The framework is used by initializing and running a master.
type Master struct {
	BaseDir string

	input   Input
	workers [][]worker
	output  Output
}

//The user must set the input, supplying at least the GenInput function.
func (m *Master) SetInput(input Input) {
	if input.Distribute == nil {
		input.Distribute = MakeRoundRobinDistributor()
	}
	m.input = input
}

//The user must set each layer, specifying the number of goroutines to use and 
//supplying at least the Map and Reduce functions.
func (m *Master) SetLayer(num int, job Job) {
	var mapLayer []worker
	var redLayer []worker
	for i := 0; i < num; i++ {
		mapDistribute := job.MapDistribute
		if mapDistribute == nil {
			mapDistribute = MakeHashDistributor()
		}
		mapLayer = append(mapLayer, &mapWorker{
			distribute: mapDistribute,
			Map:        job.Map,
		})
		redDistribute := job.RedDistribute
		if redDistribute == nil {
			redDistribute = MakeRoundRobinDistributor()
		}
		redLayer = append(redLayer, &redWorker{
			distribute: redDistribute,
			Reduce:     job.Reduce,
		})
	}
	m.workers = append(m.workers, mapLayer)
	m.workers = append(m.workers, redLayer)
}

//The user must set the output, supplying at least the GenOutput function.
func (m *Master) SetOutput(output Output) {
	if output.InitOutput == nil {
		output.InitOutput = func(param string) {}
	}
	if output.EndOutput == nil {
		output.EndOutput = func() {}
	}
	m.output = output
}

//Build builds the channels that the goroutines will use to communicate.
func (m *Master) Build() {
	outChannel := make(chan [2]string, 100)

	var channels [][]chan [2]string
	for i := 0; i < len(m.workers); i++ {
		var newChannels []chan [2]string
		for j := 0; j < len(m.workers[i]); j++ {
			newChannels = append(newChannels, make(chan [2]string, 100))
		}
		channels = append(channels, newChannels)
	}
	for j := 0; j < len(m.workers[0]); j++ {
		m.workers[0][j].init(1, channels[0][j], channels[1])
	}
	for i := 1; i < len(m.workers)-1; i++ {
		for j := 0; j < len(m.workers[i]); j++ {
			m.workers[i][j].init(len(m.workers[i-1]), channels[i][j], channels[i+1])
		}
	}
	last := len(m.workers) - 1
	for j := 0; j < len(m.workers[last]); j++ {
		m.workers[last][j].init(len(m.workers[last-1]), channels[last][j], []chan [2]string{outChannel})
	}

	m.input.init(channels[0])

	m.output.init(len(m.workers[last]), outChannel)

	m.output.numUpstream = len(m.workers[last])
	m.output.inChannel = outChannel
	m.output.endChannel = make(chan int)
}

//Start starts all of the goroutines and waits for the output.
func (m *Master) Start() int {
	go m.input.run()
	for _, workers := range m.workers {
		for _, worker := range workers {
			go worker.run()
		}
	}
	go m.output.run()
	return <-m.output.endChannel
}

//Run calls Build() and then Start()
func (m *Master) Run() int {
	m.Build()
	return m.Start()
}
