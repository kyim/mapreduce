package main

import (
	"flag"
	"fmt"
	. "mapreduce/datatypes"
	dg "mapreduce/examples/directed_graph"
	ii "mapreduce/examples/inverted_index"
	wi "mapreduce/webinterface"
	"os"
	"strings"
)

var web = flag.Bool("w", false, "run web interface")

func main() {
	flag.Parse()

	if *web {
		//Optionally specify default base directory
		//If not set, the defaults in the web interface will be used
		if len(flag.Args()) >= 1 {
			wi.Base = flag.Args()[0]
		}
		
		//Optionally specify the default input/output locations
		//The default base directory must be set first
		//The input/output locations must be set together or not at all
		if len(flag.Args()) >= 3 {
			wi.Input = flag.Args()[1]
			wi.Output = flag.Args()[2]
		}
		
		//Register jobs by name. This populates the drop-down menu
		wi.RegisterJob("Inverted Index", Job{Map: ii.MapIndex, Reduce: ii.ReduceIndex})
		wi.RegisterJob("Directed Graph 1", Job{Map: dg.MapGraph1, Reduce: dg.ReduceGraph1,
			RedDistribute: MakeHashDistributor()})
		wi.RegisterJob("Directed Graph 2", Job{Map: dg.MapGraph2, Reduce: dg.ReduceGraph2})
		
		//Hostname and port can be changed freely
		fmt.Printf("Web interface listening on: %s:%d\n", wi.Hostname, wi.Port)
		
		//Start the web interface
		wi.Run()
	} else {

		//The default base directory is the current working directory
		baseDir, err := os.Getwd()
		if err != nil {
			baseDir = "/"
		}
			
		//Default input/output locations, relative to the base directory
		input := "input/"
		output := "output.txt"
		
		//Set base directory
		if len(flag.Args()) >= 1 {
			baseDir = flag.Args()[0]
		}
		
		//Set input/output, similar to the web interface above
		if len(flag.Args()) >= 3 {
			input = flag.Args()[1]
			output = flag.Args()[2]
		}
		
		if !strings.HasSuffix(baseDir, "/") {
			baseDir += "/"
		}
		inputLoc := baseDir + input
		outputLoc := baseDir + output

		master := Master{BaseDir: baseDir}
		master.SetInput(Input{Param: inputLoc, GenInput: FileInput})

		master.SetLayer(10, Job{Map: dg.MapGraph1, Reduce: dg.ReduceGraph1,
			RedDistribute: MakeHashDistributor()})
		master.SetLayer(10, Job{Map: dg.MapGraph2, Reduce: dg.ReduceGraph2})

		//These two "versions" are identical, they serve only to demonstrate
		//two different ways of specifying output, using structs or closures
		version := 2
		if version == 1 {
			//Struct
			output := FileOutputStruct{}
			master.SetOutput(Output{Param: outputLoc, InitOutput: output.InitFileOutput, GenOutput: output.GenFileOutput, EndOutput: output.EndFileOutput})
		} else if version == 2 {
			//Closure. See datatypes/builtins.go
			i, g, e := MakeFileOutput()
			master.SetOutput(Output{Param: outputLoc, InitOutput: i, GenOutput: g, EndOutput: e})
		}

		result := master.Run()
		fmt.Printf("%d results written\n", result)
	}

}
