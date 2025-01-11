package main

import (
	"flag"
	"log"
	"net/http"
)

func main() {

	port := flag.String("port", ":8080", "port to serve")
	flag.Parse()

	mux := http.NewServeMux()
	v1 := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Her vil det komme dokumentasjon og kanskje en web app"))
	})

	v1.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hei verden"))
	})

	mux.Handle("/v1/", http.StripPrefix("/v1", v1))

	server := http.Server{
		Addr:    *port,
		Handler: mux,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal("Server shut down:", err)
	}
}
