package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/jimen0/golang-runner/scheduler"
	"google.golang.org/protobuf/proto"
)

// handleURL receives and URL and parses it.
func handleURL() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		var j scheduler.Job
		if err := proto.Unmarshal(b, &j); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		res := scheduler.Result{Id: "golang"}
		u, err := url.Parse(j.Address)
		if err != nil {
			res.Error = err.Error()
		}
		if u != nil {
			res.Value = fmt.Sprintf("Scheme=%s; Host=%s; Path=%s;", u.Scheme, u.Host, u.EscapedPath())
			if u.User != nil {
				res.Value += fmt.Sprintf(" User=%s;", u.User.String())
			}
		}

		b, err = proto.Marshal(&res)
		if err != nil {
			log.Printf("could not marshal result: %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		_, err = w.Write(b)
		if err != nil {
			log.Printf("could not write response: %v", err)
		}
	}
}

func main() {
	http.HandleFunc("/", handleURL())

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Listening on port %s", port)
	err := http.ListenAndServe(":"+port, http.DefaultServeMux)
	if err != nil {
		log.Fatalf("could not listen: %v", err)
	}
}
