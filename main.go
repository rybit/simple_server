package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

/*
 Usage: go run main.go [-p <port>] path[:status[:content]]
*/
func main() {
	var port int
	flag.IntVar(&port, "port", 9090, "the port to run on")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		fmt.Println("usage: [-port <port>] path[:status[:content]]")
		os.Exit(1)
	}

	mux := http.NewServeMux()
	for _, a := range args {
		path, status, content := parsePathDef(a)
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}

		mux.Handle(path, buildHandler(path, status, content))
	}
	log.Println("Starting to listen on port", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux); err != nil {
		fmt.Println("Failed to start server:", err)
		os.Exit(1)
	}
}

func buildHandler(path string, status int, content []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s: responding with status %d\n", path, status)
		w.WriteHeader(status)
		if _, err := w.Write([]byte(content)); err != nil {
			fmt.Printf("error writing response: %+v\n", err)
		}
	}
}

func parsePathDef(raw string) (string, int, []byte) {
	parts := strings.SplitN(raw, ":", 3)
	if len(parts) == 1 {
		return parts[0], http.StatusOK, []byte{}
	}

	code, err := strconv.Atoi(parts[1])
	if err != nil {
		panic("failed to split '" + parts[1] + "' into a number: " + err.Error())
	}

	if len(parts) == 2 {
		return parts[0], code, []byte{}
	}

	if strings.HasPrefix(parts[2], "@") {
		body, err := ioutil.ReadFile(parts[2][1:])
		if err != nil {
			panic("failed to open '" + parts[2][1:] + "': " + err.Error())
		}

		return parts[0], code, body
	}

	return parts[0], code, []byte(parts[2])
}
