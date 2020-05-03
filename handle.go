package differer

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/jimen0/differer/scheduler"
	"google.golang.org/protobuf/proto"
)

const (
	maxInputCount  = 4096
	maxInputLength = 128
)

type input struct {
	Addrs []string `json:"addresses"`
}

type output struct {
	Results []addressResult `json:"results"`
}

type addressResult struct {
	Runner string            `json:"runner"`
	Input  string            `json:"string"`
	Output *scheduler.Result `json:"outputs"`
}

// HandleInput creates a handler capable of sharing tasks with the runners.
func HandleInput(runners []Runner) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/differer" {
			http.NotFound(w, r)
			return
		}

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var i input
		if err := json.NewDecoder(r.Body).Decode(&i); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Bad schema")
			return
		}

		if len(i.Addrs) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "No addresses were received")
			return
		}

		if len(i.Addrs) > maxInputCount {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Only %d addresses can be passed at once", maxInputCount)
			return
		}

		for _, addr := range i.Addrs {
			if addr == "" {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprint(w, "Empty addresses are not allowed")
				return
			}

			if len(addr) > maxInputLength {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintf(w, "Max supported address length is %d", maxInputLength)
				return
			}
		}

		tasks := len(i.Addrs) * len(runners)
		log.Printf("Got %d addresses for %d runners. Total tasks is %d", len(i.Addrs), len(runners), tasks)

		var wg sync.WaitGroup
		wg.Add(tasks)

		results := make(chan addressResult, tasks)
		for _, addr := range i.Addrs {
			log.Printf("Creating job for %s", addr)
			j := &scheduler.Job{Address: addr}

			b, err := proto.Marshal(j)
			if err != nil {
				log.Printf("could not build task for address %q: %v", j.Address, err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			go func(addr string, data []byte) {
				for _, rn := range runners {
					ar := addressResult{Runner: rn.GetName(), Input: addr}
					res, err := rn.Run(r.Context(), data)
					if err != nil {
						log.Printf("Error while calling %s runner for %q: %v", rn.GetName(), addr, err)
						ar.Output = res
						results <- ar
						wg.Done()
						continue
					}
					ar.Output = res
					results <- ar
					wg.Done()
				}
			}(j.Address, b)
		}

		wg.Wait()

		var out output
		out.Results = make([]addressResult, 0, tasks)
		for i := 0; i < tasks; i++ {
			v := <-results
			out.Results = append(out.Results, v)
		}
		close(results)

		b, err := json.Marshal(out)
		if err != nil {
			log.Printf("Could not marshal results: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "Something went wrong internally.")
			return
		}

		w.Header().Add("content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%s", b)
	}
}
