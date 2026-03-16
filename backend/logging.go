package checkered

import (
	"log"
	"net"
	"strings"
)

// log the address the server is running on
func LogAddress(addr string, serverType string) {
	if strings.HasPrefix(addr, ":") {
		// port provided, so determine the local IP address
		// source: https://gosamples.dev/local-ip-address/
		conn, err := net.Dial("udp", "8.8.8.8:80")
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()

		localAddress := conn.LocalAddr().(*net.UDPAddr)
		log.Printf("%s running on %s%s\n", serverType, localAddress.IP, addr)
	} else {
		log.Printf("%s running on %s\n", serverType, addr)
	}
}
