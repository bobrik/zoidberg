package application

import (
	"encoding/json"
	"fmt"
)

// App represents a single application
type App struct {
	Name    string            `json:"name"`
	Servers []Server          `json:"servers"`
	Meta    map[string]string `json:"meta"`
}

// Apps is a map of app names to app instances
type Apps map[string]App

// Server is an instance of an app
type Server struct {
	Host    string `json:"host"`
	Port    int    `json:"port"`
	Ports   []int  `json:"ports"`
	Version string `json:"version"`
}

func (s Server) String() string {
	p, _ := json.Marshal(s.Ports)
	return fmt.Sprintf("%s%s", s.Host, p)
}
