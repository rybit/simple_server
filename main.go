package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

type configuration struct {
	Routes map[string]struct {
		Headers map[string]string
		Status  int
		Payload string
		Ref     string
	}
}

func main() {
	var port int
	flag.IntVar(&port, "port", 9090, "the port to run on")
	flag.Parse()

	if len(flag.Args()) == 0 {
		fmt.Println("usage: [-port <port>] config.json")
		os.Exit(1)
	}

	bs, err := ioutil.ReadFile(flag.Arg(0))
	if err != nil {
		log.Fatalf("failed to open the config file %s", flag.Arg(0))
	}

	var config configuration
	if err := json.Unmarshal(bs, &config); err != nil {
		log.Fatalf("failed to unmarshal the config: %+v\n", err)
	}
	if len(config.Routes) == 0 {
		log.Fatalln("no routes defined")
	}

	mux := http.NewServeMux()
	for path, a := range config.Routes {
		var body []byte
		if a.Ref != "" {
			data, err := ioutil.ReadFile(a.Ref)
			if err != nil {
				log.Fatalf("error reading referenced file: %+v\n", err)
			}
			body = data
		}
		if a.Payload != "" {
			body = []byte(a.Payload)
		}
		status := a.Status
		if status == 0 {
			status = http.StatusOK
		}
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		log.Printf("Registering path: %s", path)
		mux.Handle(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("%s: responding with status %d\n", path, a.Status)
			for k, v := range a.Headers {
				w.Header().Set(k, v)
			}

			w.WriteHeader(status)
			if _, err := w.Write(body); err != nil {
				fmt.Printf("error writing response: %+v\n", err)
			}
		}))

	}
	log.Println("Starting to listen on port", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux); err != nil {
		fmt.Println("Failed to start server:", err)
		os.Exit(1)
	}
}
