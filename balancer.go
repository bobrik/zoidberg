package zoidberg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Balancer struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

type balancerState struct {
	Apps  Apps  `json:"apps"`
	State State `json:"state"`
}

func (b Balancer) update(name string, apps Apps, state State) error {
	body, err := json.Marshal(balancerState{
		Apps:  apps,
		State: state,
	})

	if err != nil {
		return err
	}

	u := fmt.Sprintf("http://%s:%d/state/%s", b.Host, b.Port, name)
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
