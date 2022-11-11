package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func bootServer(opArg string) {
	const (
		node   = "node"
		index  = "express/index.js"
		rust   = "rust"
		golang = "go"
		godir  = "golang/main.go"
	)

	var args []string
	switch opArg {
	case golang:
		args = append(args, "run", godir)
	case node:
		args = append(args, index)
	case rust:
		os.Chdir(rust)
		opArg = "cargo"
		args = append(args, "run", "--release")
	}

	go func() {
		cmd := exec.Command(opArg, args...)
		err := cmd.Run()
		if err != nil {
			fmt.Println(fmt.Errorf(
				"err in server-booting subcommand: %w", err,
			))
		}
	}()
}

func killServer(opArg string) {
	psStrs := getPsOf(opArg)
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

// TODO: Currently busted
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

	opArg := ""
	rps := 180
	bypass := false

	// verbose opt
	printServOutputs := false
	args := os.Args[1:]
	if len(args) > 0 {
		for _, arg := range args {
			if arg == "-v" {
				printServOutputs = true
				continue
			} else if arg == "-by" {
				bypass = true
			}

			if strings.Contains(arg, "rps") {
				split := strings.Split(arg, "=")
				n, err := strconv.Atoi(split[1])
				if err != nil {
					fmt.Println(err)
					continue
				}
				rps = n
			}

			opArg = arg
		}
	}
	if len(opArg) == 0 {
		fmt.Println("no valid args")
		os.Exit(0)
	}

	delay := 400 * time.Millisecond
	xmlBytes := getXml()

	if !bypass {
		bootServer(opArg)
		switch opArg {
		case "rust":
			opArg = "rustyserver"
			delay = 6 * time.Second
			fmt.Println("rust arg provided; setting compile delay to 6 seconds; might not be enough")
		}
		defer killServer(opArg)

		// time for node to boot / rust to compile
		time.Sleep(delay)
	}

	fmt.Printf("targeting an rps of: \x1b[34m%v\x1b[0m\n", rps)
	go hitRps(rps, xmlBytes, printServOutputs)
	for {
		getCPULoadOf(opArg)
		time.Sleep(500 * time.Millisecond)
	}
}
