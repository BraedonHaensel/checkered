package main

import (
	"flag"
	"log"
	"net/http"

	Checkered "github.com/akeuben/checkered"
)

var addr = flag.String("addr", ":5000", "http service address")
var nameServerUrl = flag.String("ns", "http://localhost:9000", "full Name Server URL")

func main() {
	flag.Parse()
	server := Checkered.InitServer()
	go server.ServerLoop()
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		Checkered.ServeWs(server, w, r)
	})

	// Register with the Name Server
	url := Checkered.GetFullUrl(*addr)
	Checkered.SendRegistrationRequest(url, *nameServerUrl+"/register/game-server")

	log.Println("Game Server running on", url)
	err := http.ListenAndServe(*addr, Checkered.CORSMiddleware(http.DefaultServeMux))
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
