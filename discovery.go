package zoidberg

import (
	"github.com/bobrik/zoidberg/application"
	"github.com/bobrik/zoidberg/balancer"
)

// Discovery is a current known state of the world: load balancers and apps
type Discovery struct {
	Balancers []balancer.Balancer `json:"balancers"`
	Apps      application.Apps    `json:"apps"`
}
