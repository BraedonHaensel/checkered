package main

import (
	"flag"
	"log"
	"net/http"

	Checkered "github.com/akeuben/checkered"
)

var (
	addr          = flag.String("addr", ":5000", "http service address")
	nameServerURL = flag.String("ns", "http://localhost:9000", "full Name Server URL")
)

func main() {
	flag.Parse()
	server := Checkered.InitServer(*nameServerURL)
	go server.ServerLoop()
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		Checkered.ServeWs(server, w, r)
	})

	// Register with the Name Server
	url := Checkered.GetFullURL(*addr)
	log.Println("Game Server running on", url)
	server.Register(url)

	err := http.ListenAndServe(*addr, Checkered.CORSMiddleware(http.DefaultServeMux))
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
