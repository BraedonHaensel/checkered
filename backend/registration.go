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

type registerServerResponse struct {
	ID int `json:"id"`
}

// Gets the fully qualified URL for a server address.
func GetFullURL(addr string) string {
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

// Sends a registration request for a server to the Name Server. Returns the
// assigned ID for the server.
func SendRegistrationRequest(serverUrl, registrationUrl string) int {
	// Create the registration request
	body := fmt.Appendf(nil, `{"url": "%s"}`, serverUrl)
	res, err := http.Post(registrationUrl, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Fatal("Name Server registration failed: ", err)
	}
	defer res.Body.Close()

	// Check the status code to see if the request succeeded
	if res.StatusCode != http.StatusCreated {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			log.Fatalf("Name Server registration failed with status code %d. Failed "+
				"to parse the response body: %v", res.StatusCode, err)
		}
		log.Fatalf("Name Server registration failed with status code %d: %s", res.StatusCode, string(body))
	}

	// Parse the registered ID from the response
	data, err := ParseJsonResponseData[registerServerResponse](res)
	if err != nil {
		log.Fatal("Name Server registration failed: ", err)
	}
	return data.ID
}
