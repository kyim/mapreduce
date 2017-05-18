//Package webinterface provides a simple framework for defining and running
//simple web interfaces to build and run mapreduce jobs
package webinterface

import (
	"bytes"
	"fmt"
	d "mapreduce/datatypes"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var jobMap map[string]d.Job = make(map[string]d.Job)

//Base, Input, Output, and Num are defaults that can be edited to change the
//defaults that will be presented to the user. Their values are unimportant
//if the user wants to set them themself every time instead.
var Base, Input, Output string = getwd(), "input/", "output.txt"
var Num int = 10
//Hostname and Port determine the binding of the server
var Hostname string =  "localhost"
var Port int = 8080

func getwd() string {
	wd, err := os.Getwd()
	if err != nil {
		return "/"
	}
	return wd
}

//Jobs must be registered by name in order to be accessible from the web
//interface
func RegisterJob(name string, job d.Job) {
	jobMap[name] = job
}

//Run() runs http.ListenAndServe for the current hostname and port
func Run() error {
	http.HandleFunc("/", handler)
	http.HandleFunc("/submit", submit)
	return http.ListenAndServe(fmt.Sprintf("%s:%d", Hostname, Port), nil)
}

func submit(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		http.Error(w, fmt.Sprintf("Could not parse request: %v", err), 500)
	}

	if r.PostForm["layer"] == nil || r.PostForm["num"] == nil ||
		len(r.PostForm["layer"]) != len(r.PostForm["num"]) {
		http.Error(w, "Malformed request", 400)
	}

	baseDir := r.FormValue("baseDir")
	if baseDir == "" {
		http.Redirect(w, r, "/", 307)
		return
	}

	if !strings.HasSuffix(baseDir, "/") {
		baseDir += "/"
	}
	stdIn := r.FormValue("stdIn")
	stdOut := r.FormValue("stdOut")

	master := d.Master{BaseDir: baseDir}

	if stdIn != "" {
		master.SetInput(d.Input{Param: "", GenInput: d.StdInput})
	} else {
		inputLoc := baseDir + r.FormValue("input")
		master.SetInput(d.Input{Param: inputLoc, GenInput: d.FileInput})
	}

	for i, l := range r.PostForm["layer"] {
		num, err := strconv.Atoi(r.PostForm["num"][i])
		if err != nil {
			http.Error(w, fmt.Sprintf("Illegal 'num': %s", r.PostForm["num"][i]), 400)
		}

		master.SetLayer(num, jobMap[l])
	}

	if stdOut != "" {
		master.SetOutput(d.Output{Param: "", GenOutput: d.StdOutput})
	} else {
		outputLoc := baseDir + r.FormValue("output")
		i, g, e := d.MakeFileOutput()
		master.SetOutput(d.Output{Param: outputLoc, InitOutput: i, GenOutput: g, EndOutput: e})
	}

	go master.Run()

	http.Redirect(w, r, "/", 307)

}
func handler(w http.ResponseWriter, r *http.Request) {
	var buffer bytes.Buffer
	for name := range jobMap {
		buffer.WriteString("<option value=\"")
		buffer.WriteString(name)
		buffer.WriteString("\">")
		buffer.WriteString(name)
		buffer.WriteString("</option>\n")
	}

	output := `<html><head>
<script>
var num = 1;
function addLayer() {
	var parent = document.getElementById('layers');
	var child = document.getElementById('layer').cloneNode(true);
	child.getElementsByClassName("header")[0].innerHTML += num+":";
	
	num++;
	
	parent.appendChild(child);
	return false;
}
function check(name, box) {
	if (box.checked) {
        document.getElementById(name).style.display = 'none';
	} else {
        document.getElementById(name).style.display = 'block';
	}
	return false;
}
</script>
</head><body>

<div id="hide" style="display: none;">
<div id="layer">
<h class="header">Layer </h>
<select name="layer">` +
		buffer.String() +
		`</select>
<br>  
Number of workers:<br>
<input type="text" name="num" value="` +
		strconv.Itoa(Num) +
		`">
<br>
</div>
</div>

<form action="/submit" method="post">
Base Directory:<br>
<input type="text" name="baseDir" value="` +
		Base +
		`">
<br>
<div id="inputBox">
Input:<br>
<input type="text" name="input" value="` +
		Input +
		`">
<br>
</div>

<input type="checkbox" name="stdIn" value="stdIn"
onclick="check('inputBox', this)">
 Read from standard in<br>
 
<div id="outputBox">
Output:<br>
<input type="text" name="output" value="` +
		Output +
		`">
<br>
</div>

<input type="checkbox" name="stdOut" value="stdOut"
onclick="check('outputBox', this)">
 Print to standard out<br>
 
<div id="layers">
</div>
<button type="button"
onclick="addLayer()">
Add New Layer</button>
<br>
<input type="submit" value="Submit">
</form>
</body></html>`
	fmt.Fprintln(w, output)
}
