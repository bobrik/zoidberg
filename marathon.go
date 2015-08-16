package zoidberg

import (
	"log"
	"net/url"

	"github.com/gambol99/go-marathon"
)

type MarathonDiscoverer struct {
	m        marathon.Marathon
	balancer string
	static   []Balancer
}

func NewMarathonDiscoverer(m marathon.Marathon, b string) MarathonDiscoverer {
	return MarathonDiscoverer{
		m:        m,
		balancer: b,
	}
}

func (m *MarathonDiscoverer) SetStaticBalancers(b []Balancer) {
	m.static = b
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
	if m.static != nil {
		return m.static, nil
	}

	mv := url.Values{}
	mv.Set("embed", "apps.tasks")
	mv.Set("label", "zoidberg_balancer_for=="+m.balancer)
	ma, err := m.m.Applications(mv)
	if err != nil {
		return nil, err
	}

	balancers := []Balancer{}
	for _, app := range ma.Apps {
		if len(app.Ports) == 0 {
			log.Printf("app %s has no ports\n", app.ID)
			continue
		}

		for _, task := range app.Tasks {
			balancers = append(balancers, Balancer{
				Host: task.Host,
				Port: task.Ports[0],
			})
		}
	}

	return balancers, nil
}

func (m MarathonDiscoverer) apps() (Apps, error) {
	mv := url.Values{}
	mv.Set("embed", "apps.tasks")
	mv.Set("label", "zoidberg_balanced_by=="+m.balancer)
	ma, err := m.m.Applications(mv)
	if err != nil {
		return nil, err
	}

	apps := map[string]App{}

	for _, a := range ma.Apps {
		name := a.Labels["zoidberg_app_name"]
		if name == "" {
			log.Printf("app %s has no label zoidberg_app_name\n", a.ID)
			continue
		}

		version := a.Labels["zoidberg_app_version"]
		if version == "" {
			log.Printf("app %s has no label zoidberg_app_version\n", a.ID)
			continue
		}

		app := apps[name]
		if app.Name == "" {
			app.Name = name
			app.Servers = []Server{}
		}

		for _, task := range a.Tasks {
			healthy := true
			for _, check := range task.HealthCheckResult {
				if check == nil {
					continue
				}

				if !check.Alive {
					healthy = false
					break
				}
			}

			if !healthy {
				continue
			}

			app.Servers = append(app.Servers, Server{
				Version: version,
				Host:    task.Host,
				Port:    task.Ports[0],
				Ports:   task.Ports,
			})
		}

		apps[name] = app
	}

	return apps, nil
}
