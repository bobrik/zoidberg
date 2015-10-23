package balancer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/bobrik/zoidberg/application"
	"github.com/bobrik/zoidberg/state"
)

// Balancer represents a load balancer
type Balancer struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

// State represents load balancer's state:
// known applications and their respective versions
type State struct {
	Apps  application.Apps `json:"apps"`
	State state.State      `json:"state"`
}

// Update updates load balancer's state
func (b Balancer) Update(name string, apps application.Apps, state state.State) error {
	body, err := json.Marshal(State{
		Apps:  apps,
		State: state,
	})

	if err != nil {
		return err
	}

	c := http.Client{
		Timeout: time.Second * 5,
	}

	u := fmt.Sprintf("http://%s:%d/state/%s", b.Host, b.Port, name)
	resp, err := c.Post(u, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}

	return resp.Body.Close()
}

// String returns load balancer's location string representation
func (b Balancer) String() string {
	return fmt.Sprintf("%s:%d", b.Host, b.Port)
}
