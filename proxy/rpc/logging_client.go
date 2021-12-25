package rpc

import (
	"encoding/json"
	"log"
)

type LoggingClient struct {
	Client
	serverName string
}

func (c *LoggingClient) Call(serviceMethod string, request interface{}, response interface{}) error {
	marshaledRequest, _ := json.Marshal(request)
	log.Printf("%s: <-- %s %s\n", c.serverName, serviceMethod, string(marshaledRequest))
	err := c.Client.Call(serviceMethod, request, response)
	if err != nil {
		log.Printf("%s: --> %s error %v\n", c.serverName, serviceMethod, err)
		return err
	}
	marshaledResponse, _ := json.Marshal(response)
	log.Printf("%s: --> %s %s\n", c.serverName, serviceMethod, string(marshaledResponse))
	return nil
}

func NewLoggingClient(serverName string, wrappedClient Client) *LoggingClient {
	return &LoggingClient{
		Client:     wrappedClient,
		serverName: serverName,
	}
}
