package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

// hardcoding for now, will put in opts later
const node = "node"

func bootNode() {
	go func() {
		cmd := exec.Command(node, "express/index.js")
		err := cmd.Run()
		if err != nil {
			fmt.Println(err)
		}
	}()
}

func killNode() {
	psStrs := getPsOf(node)
	pid := psStrs[1]
	cmd := exec.Command("kill", pid)
	err := cmd.Run()
	if err != nil {
		panic(err)
	}
}

// Calls `ps` for procname for cpu load / pid.
// Assumes user is not running other node processes.
// Am too lazy to write to simply grab last-spun node.
func getPsOf(procname string) []string {
	cmd := exec.Command("ps", "au")
	bytes, err := cmd.Output()
	if err != nil {
		fmt.Println(procname, "err:", err)
		return []string{}
	}

	output := string(bytes)
	lines := strings.Split(output, "\n")
	for _, ln := range lines {
		if strings.Contains(ln, procname) {
			split := strings.Split(ln, " ")
			cleanSplit := []string{}
			for _, s := range split {
				if len(s) != 0 {
					cleanSplit = append(cleanSplit, s)
				}
			}
			return cleanSplit
		}
	}
	return []string{}
}

func hitRps(reqsPerSecond int, xmlFile []byte, printOut bool) {
	delay := time.Second / time.Duration(reqsPerSecond)
	for {
		xmlRdr := bytes.NewReader(xmlFile)
		res, err := http.Post("http://localhost:3030/", "text/xml", xmlRdr)
		if err != nil {
			fmt.Println("err:", err)
		}
		if printOut {
			bytes, _ := io.ReadAll(res.Body)
			fmt.Println(string(bytes))
		}
		time.Sleep(delay)
	}
}

func getCPULoadOf(procname string) {
	ps := getPsOf(procname)
	// ps output fmt:
	// USER PID %CPU %MEM VSZ RSS TTY STAT START TIME COMMAND ARG1
	// we want   ^
	cpu := ps[2]
	fmt.Printf("load of %s: %v\n", procname, cpu)
}

func getXml() []byte {
	f, err := os.Open("control/junk.xml")
	if err != nil {
		panic(err)
	}
	bytes, _ := io.ReadAll(f)
	return bytes
}

func main() {

	// verbose opt
	printServOutputs := false
	args := os.Args[1:]
	if len(args) > 0 {
		if args[0] == "-v" {
			printServOutputs = true
		}
	}

	xmlBytes := getXml()
	bootNode()
	defer killNode()
	time.Sleep(200 * time.Millisecond) // time for node to boot? get errors w/o this
	rps := 180
	fmt.Printf("targeting an rps of: \x1b[34m%v\x1b[0m\n", rps)
	go hitRps(rps, xmlBytes, printServOutputs)
	for {
		getCPULoadOf(node)
		time.Sleep(500 * time.Millisecond)
	}
}
