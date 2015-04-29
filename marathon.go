package zoidberg

import (
	"fmt"
	"strings"

	"github.com/gambol99/go-marathon"
)

type MarathonDiscoverer struct {
	m        marathon.Marathon
	balancer string
	groups   []string
}

func NewMarathonDiscoverer(m marathon.Marathon, balancer string, groups []string) MarathonDiscoverer {
	return MarathonDiscoverer{
		m:        m,
		balancer: balancer,
		groups:   groups,
	}
}

func (m MarathonDiscoverer) Discover() (Discovery, error) {
	discovery := Discovery{}

	balancers, err := m.balancers()
	if err != nil {
		return discovery, err
	}

	apps, err := m.apps()
	if err != nil {
		return discovery, err
	}

	discovery.Balancers = balancers
	discovery.Apps = apps

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

func (m MarathonDiscoverer) apps() (Apps, error) {
	apps := map[string]App{}

	for _, group := range m.groups {
		app, err := m.app(group)
		if err != nil {
			return nil, err
		}

		apps[app.Name] = app
	}

	return apps, nil
}

func (m MarathonDiscoverer) app(group string) (App, error) {
	g, err := m.m.Group(group)
	if err != nil {
		return App{}, err
	}

	application := App{
		Servers: []Server{},
	}

	for _, app := range g.Apps {
		app, err := m.m.Application(app.ID)
		if err != nil {
			return App{}, err
		}

		// first app contributes name and port, they should be the same anyway
		if application.Name == "" {
			if name, ok := app.Labels["zoidberg_app_name"]; ok {
				application.Name = name
				application.Port = app.Ports[0] // TODO: maybe all of them?
			}
		}

		version := strings.TrimPrefix(app.ID, g.ID)

		for _, task := range app.Tasks {
			application.Servers = append(application.Servers, Server{
				Version: version,
				Host:    task.Host,
				Port:    task.Ports[0], // TODO: maybe all of them?
			})
		}
	}

	if application.Name == "" {
		return App{}, fmt.Errorf("application name not discovered for %q", group)
	}

	return application, nil
}
