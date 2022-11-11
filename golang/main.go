package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type fruit struct {
	fruitType string
	price     string
}

func parseXml(w http.ResponseWriter, rq *http.Request) {
	bytes, err := io.ReadAll(rq.Body)
	if err != nil {
		panic(err)
	}

	strs := string(bytes)
	lines := strings.Split(strs, "\n")

	responseFruit := fruit{
		fruitType: "",
		price:     "",
	}

bodyparse:
	for _, ln := range lines {
		seenFruit := false
		if strings.Contains(ln, "Strawberry") {
			idx := 0
			for i, rn := range ln {
				if rn == 'S' {
					idx = i
				} else if rn == '<' && idx > 0 {
					responseFruit.fruitType = ln[idx:i]
					break
				}
			}
		} else if seenFruit {
			idx := 0
			for i, rn := range ln {
				if rn == '>' {
					idx = i
				} else if rn == '<' && idx > 0 {
					responseFruit.price = ln[idx:i]
					break bodyparse
				}
			}
		}
	}

	jsonBytes, err := json.Marshal(responseFruit)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(responseFruit)
	w.Write(jsonBytes)
}

func main() {
	http.HandleFunc("/", parseXml)
	http.ListenAndServe(":3030", nil)
}
