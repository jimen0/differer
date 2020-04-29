package main

import (
	"log"
	"net/http"
	"os"

	"github.com/jimen0/differer"
)

func main() {
	f, err := os.Open(os.Getenv("DIFFERER_CONFIG"))
	if err != nil {
		log.Printf("could not read config: %v", err)
		return
	}
	defer f.Close()

	cfg, err := differer.ReadConfig(f)
	if err != nil {
		log.Printf("could not read config: %v", err)
		return
	}

	var runners []differer.Runner
	for name, addr := range cfg.Runners {
		runners = append(runners, &differer.CloudRunner{
			Name:    name,
			Service: addr,
			Client:  &http.Client{Timeout: cfg.Timeout},
		})
	}

	http.HandleFunc("/differer", differer.HandleInput(runners))
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Printf("could not listen: %v", err)
		return
	}
}
