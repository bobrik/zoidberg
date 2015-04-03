package explorer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Balancer struct {
	Host string `json:"host"`
	Port int `json:"port"`
}

func (b Balancer) announce(location ExplorerLocation) error {
	body, err := json.Marshal(location)
	if err != nil {
		return err
	}

	u := fmt.Sprintf("http://%s:%d/explorer_location", b.Host, b.Port)
	resp, err := http.Post(u, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}

	resp.Body.Close()

	return nil
}

func (b Balancer) updateUpstreams(upstreams []Upstream) error {
	body, err := json.Marshal(upstreams)
	if err != nil {
		return err
	}

	log.Printf("sending to %s: %s\n", b, string(body))

	u := fmt.Sprintf("http://%s:%d/upstreams", b.Host, b.Port)
	resp, err := http.Post(u, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}

	resp.Body.Close()

	return nil
}

func (b Balancer) String() string {
	return fmt.Sprintf("%s:%d", b.Host, b.Port)
}
