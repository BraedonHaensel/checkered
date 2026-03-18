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

// Each registered server has an ID and a URL
type Server struct {
	ID  int    `json:"id"`
	URL string `json:"url"`
}

type serverListResponse []Server
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
func SendRegistrationRequest(serverUrl, registrationUrl string) (int, error) {
	// Create the registration request
	body := fmt.Appendf(nil, `{"url": "%s"}`, serverUrl)
	res, err := http.Post(registrationUrl, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return 0, fmt.Errorf("Name Server registration failed: %w", err)
	}
	defer res.Body.Close()

	// Check the status code to see if the request succeeded
	if res.StatusCode != http.StatusCreated {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return 0, fmt.Errorf("Name Server registration failed with status code %d. Failed "+
				"to parse the response body: %w", res.StatusCode, err)
		}
		return 0, fmt.Errorf("Name Server registration failed with status code %d: %s", res.StatusCode, string(body))
	}

	// Parse the registered ID from the response
	data, err := ParseJsonResponseData[registerServerResponse](res)
	if err != nil {
		return 0, fmt.Errorf("Name Server registration failed: %w", err)
	}
	return data.ID, nil
}

// Sends a server list request to the Name Server.
func SendServerListRequest(serverListURL string) ([]Server, error) {
	// Create the server list request
	res, err := http.Get(serverListURL)
	if err != nil {
		return nil, fmt.Errorf("server list refresh failed: %w", err)
	}
	defer res.Body.Close()

	// Parse the server list from the response
	serverList, err := ParseJsonResponseData[serverListResponse](res)
	if err != nil {
		return nil, fmt.Errorf("server list refresh failed: %w", err)
	}
	return serverList, nil
}
