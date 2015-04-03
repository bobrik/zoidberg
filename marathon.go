package explorer

import (
	"strings"

	"github.com/gambol99/go-marathon"
)

type MarathonDiscoverer struct {
	m        marathon.Marathon
	balancer string
	group    string
}

func NewMarathonDiscoverer(m marathon.Marathon, balancer, group string) MarathonDiscoverer {
	return MarathonDiscoverer{
		m:        m,
		balancer: balancer,
		group:    group,
	}
}

func (m MarathonDiscoverer) Discover() (Discovery, error) {
	discovery := Discovery{}

	balancers, err := m.balancers()
	if err != nil {
		return discovery, err
	}

	servers, err := m.servers()
	if err != nil {
		return discovery, err
	}

	discovery.Balancers = balancers
	discovery.Servers = servers

	return discovery, nil
}

func (m MarathonDiscoverer) balancers() ([]Balancer, error) {
	app, err := m.m.Application(m.balancer)
	if err != nil {
		return nil, err
	}

	balancers := make([]Balancer, len(app.Tasks))

	for i, task := range app.Tasks {
		balancers[i] = Balancer{
			Host: task.Host,
			Port: task.Ports[0],
		}
	}

	return balancers, nil
}

func (m MarathonDiscoverer) servers() (map[string][]Server, error) {
	group, err := m.m.Group(m.group)
	if err != nil {
		return nil, err
	}

	servers := map[string][]Server{}

	for _, app := range group.Apps {
		app, err := m.m.Application(app.ID)
		if err != nil {
			return nil, err
		}

		version := strings.TrimPrefix(app.ID, group.ID)
		servers[version] = make([]Server, len(app.Tasks))

		for i, task := range app.Tasks {
			servers[version][i] = Server{
				Host: task.Host,
				Port: task.Ports[0],
			}
		}
	}

	return servers, nil
}
