package client

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"golang.org/x/time/rate"
)

// TCPClient a tcp client
type TCPClient struct {
	address string
}

//GetAddress get client address
func (c *TCPClient) GetAddress() string {
	return c.address
}

// New create a new TCPClient
func New(remoteAddress string) *TCPClient {
	return &TCPClient{remoteAddress}
}

// APIClient external api client
type APIClient struct {
	address string
	client  http.Client
	limiter *rate.Limiter
}

//DoSend sedn request to external api
func (ac *APIClient) DoSend(cmd string) string {
	if ac.limiter.Allow() == false {
		return fmt.Sprintf("%v\n", "Too Many Requests.")
	}

	//connection time out
	ac.client.Timeout = time.Duration(5 * time.Second)
	resp, err := ac.client.Post(
		ac.address,
		"application/x-www-form-urlencoded",
		strings.NewReader(cmd))

	if err != nil {
		log.Println(`API error:`, err)
		if strings.LastIndexAny(err.Error(), "connection refused") > 1 {
			return fmt.Sprintf("%v\n", "Server unreachable.")
		}
		return fmt.Sprintf("%v\n", err.Error())
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(`API error:`, err)
		return fmt.Sprintf("%v\n", err.Error())
	}
	return fmt.Sprintf("%v\n", string(body))
}

// NewAPI create a new APIClient
func NewAPI(remoteAddress string) *APIClient {
	return &APIClient{remoteAddress, http.Client{}, rate.NewLimiter(2, 5)}
}
