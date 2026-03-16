package main

import (
	"flag"
	"log"
	"net/http"

	Checkered "github.com/akeuben/checkered"
)

var addr = flag.String("addr", ":5000", "http service address")

func main() {
	flag.Parse()
	server := Checkered.InitServer()
	go server.ServerLoop()
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		Checkered.ServeWs(server, w, r)
	})
	Checkered.LogAddress(*addr, "Game Server")
	err := http.ListenAndServe(*addr, Checkered.CORSMiddleware(http.DefaultServeMux))
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
