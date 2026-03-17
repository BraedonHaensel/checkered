package checkered

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
)

// Gets the fully qualified URL for a server address.
func GetFullUrl(addr string) string {
	if strings.HasPrefix(addr, ":") {
		// Port provided, so determine the local IP address
		// Source: https://gosamples.dev/local-ip-address/
		conn, err := net.Dial("udp", "8.8.8.8:80")
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()

		localAddress := conn.LocalAddr().(*net.UDPAddr)
		return "http://" + localAddress.IP.String() + addr
	} else {
		// Add "http://" to the address
		return "http://" + addr
	}
}

// Sends a registration request for a server to the Name Server.
func SendRegistrationRequest(serverUrl, registrationUrl string) {
	// Create the registration request
	body := fmt.Appendf(nil, `{"url": "%s"}`, serverUrl)
	req, err := http.NewRequest("POST", registrationUrl, bytes.NewBuffer(body))
	if err != nil {
		log.Fatal("failed to create Name Server registration request: ", err)
	}
	req.Header.Add("Content-Type", "application/json")
	
	// Send the registration request
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Fatal("Name Server registration failed: ", err)
	}
	defer res.Body.Close()
	
	// Check the status code to see if the request succeeded
	if res.StatusCode != http.StatusCreated {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			log.Fatalf("Name Server registration failed with status code %d. Failed "+
			"to parse the response body: %s", res.StatusCode, err)
		}
		log.Fatalf("Name Server registration failed with status code %d: %s", res.StatusCode, string(body))
	}
}
	