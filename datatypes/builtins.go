package datatypes

import "hash/fnv"
import "math/rand"
import "time"
import "os"
import "bufio"
import "fmt"
import "io/ioutil"
import "strconv"

func MakeRandomRoundRobinDistributor(size int) Distributor {
	count := 0
	if size > 0 {
		rand.Seed(time.Now().UnixNano())
		count = rand.Intn(size)
	}
	return func(data [2]string, channels []chan [2]string) {
		if count >= len(channels) {
			count = 0
		}
		channels[count] <- data
		count++
	}
}

func MakeRoundRobinDistributor() Distributor {
	return MakeRandomRoundRobinDistributor(0)
}

func MakeHashDistributor() Distributor {
	return func(data [2]string, channels []chan [2]string) {
		h := fnv.New32a()
		h.Write([]byte(data[0]))
		channels[int(h.Sum32())%len(channels)] <- data
	}
}

func inputErr(err error) {
	fmt.Printf("Input termination due to error: %v\n", err)
}

//FileInput reads from a file, or all of the files in a directory.
//Values are the entire line, keys are the filename and line number,
func FileInput(param string, emitter Emitter) {

	info, err := os.Stat(param)
	if err != nil {
		inputErr(err)
		return
	}

	if info.IsDir() {

		dirs, err := ioutil.ReadDir(param)
		if err != nil {
			inputErr(err)
			return
		}
		for _, f := range dirs {
			if f.Mode().IsRegular() {
				file, err := os.Open(param + "/" + f.Name())
				if err != nil {
					inputErr(err)
					return
				}

				scanner := bufio.NewScanner(file)
				for i := 0; scanner.Scan(); i++ {
					emitter.Emit(fmt.Sprintf("%s:%d", f.Name(), i), scanner.Text())
				}
				file.Close()
			}
		}
	} else {
		file, err := os.Open(param)
		if err != nil {
			inputErr(err)
		}

		scanner := bufio.NewScanner(file)
		for i := 0; scanner.Scan(); i++ {
			emitter.Emit(fmt.Sprintf("%s:%d", info.Name(), i), scanner.Text())
		}
		file.Close()
	}
}

//StdInput generates input from standard in,
func StdInput(param string, emitter Emitter) {
	scanner := bufio.NewScanner(os.Stdin)
	for i := 0; scanner.Scan(); i++ {
		emitter.Emit(strconv.Itoa(i), scanner.Text())
	}
}

func outputErr(err error) {
	fmt.Printf("Output failure due to error: %v\n", err)
}

//This struct is used to pass values between functions.
//The functions open a file, write the values to the file, then close it.
//See MakeFileOutput() for the same functionality using closures.
type FileOutputStruct struct {
	f *os.File
	w *bufio.Writer
}

func (g *FileOutputStruct) InitFileOutput(param string) {
	f, err := os.Create(param)
	if err != nil {
		outputErr(err)
		return
	}
	g.f = f
	g.w = bufio.NewWriter(f)
}
func (g *FileOutputStruct) GenFileOutput(param, key, value string) {
	fmt.Fprintln(g.w, value)
}
func (g *FileOutputStruct) EndFileOutput() {
	g.w.Flush()
	g.f.Close()
}

//This function returns three functions for handling output.
//The functions open a file, write the values to the file, then close it.
//See FileOutputStruct for the same functionality using structs.
func MakeFileOutput() (i func(param string), g func(param, key, value string), e func()) {
	var f *os.File
	var w *bufio.Writer
	i = func(param string) {
		f, err := os.Create(param)
		if err != nil {
			outputErr(err)
		}
		w = bufio.NewWriter(f)
	}
	g = func(param, key, value string) {
		fmt.Fprintln(w, value)
	}
	e = func() {
		w.Flush()
		f.Close()
	}
	return
}

//This function outputs the values to standard out,
//Since there are no init or end functions, only a single function is needed,
func StdOutput(param, key, value string) {
	fmt.Println(value)
}
