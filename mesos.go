package zoidberg

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type refinedMesosTask struct {
	Name   string
	Host   string
	Ports  []int
	Labels map[string]string
}

type mesosState struct {
	Frameworks []mesosFramework `json:"frameworks"`
	Slaves     []mesosSlave     `json:"slaves"`
}

type mesosFramework struct {
	Tasks []mesosTask `json:"tasks"`
}

type mesosTask struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	State     string         `json:"state"`
	SlaveID   string         `json:"slave_id"`
	Resources mesosResources `json:"resources"`
	Labels    []mesosLabel   `json:"labels"`
}

type mesosSlave struct {
	ID   string `json:"id"`
	Host string `json:"hostname"`
}

type mesosResources struct {
	Ports mesosPorts `json:"ports"`
}

type mesosLabel struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type mesosPorts []int

func (mp *mesosPorts) UnmarshalJSON(b []byte) error {
	p := strings.TrimFunc(string(b), func(r rune) bool {
		return r == '[' || r == ']' || r == '"'
	})

	ports := []int{}

	for _, s := range strings.Split(p, " ") {
		r := strings.Split(s, "-")
		if len(r) != 2 {
			return fmt.Errorf("expected port range to be in formate XXXX-YYYY, got %s", r)
		}

		start, err := strconv.Atoi(r[0])
		if err != nil {
			return err
		}

		end, err := strconv.Atoi(r[1])
		if err != nil {
			return err
		}

		for i := start; i <= end; i++ {
			ports = append(ports, i)
		}
	}

	*mp = mesosPorts(ports)

	return nil
}

type MesosDiscoverer struct {
	masters  []string
	balancer string
}

func NewMesosDiscoverer(masters []string, balancer string) MesosDiscoverer {
	return MesosDiscoverer{
		masters:  masters,
		balancer: balancer,
	}
}

func (m MesosDiscoverer) Discover() (Discovery, error) {
	discovery := Discovery{}

	tasks, err := m.tasks()
	if err != nil {
		return discovery, err
	}

	discovery.Balancers = m.balancers(tasks)
	discovery.Apps = m.apps(tasks)

	return discovery, nil
}

func (m MesosDiscoverer) balancers(tasks []refinedMesosTask) []Balancer {
	balancers := []Balancer{}

	for _, task := range tasks {
		if task.Labels["zoidberg_balancer_for"] == m.balancer {
			balancers = append(balancers, Balancer{
				Host: task.Host,
				Port: task.Ports[0],
			})
		}
	}

	return balancers
}

func (m MesosDiscoverer) apps(tasks []refinedMesosTask) Apps {
	apps := map[string]App{}

	for _, task := range tasks {
		if task.Labels["zoidberg_balanced_by"] != m.balancer {
			continue
		}

		name := task.Labels["zoidberg_app_name"]
		if name == "" {
			continue
		}

		version := task.Labels["zoidberg_app_version"]
		if version == "" {
			continue
		}

		app := apps[name]
		if app.Name == "" {
			port, err := strconv.Atoi(task.Labels["zoidberg_app_port"])
			if err == nil {
				app.Name = name
				app.Port = port
			}
		}

		app.Servers = append(app.Servers, Server{
			Version: version,
			Host:    task.Host,
			Port:    task.Ports[0],
		})

		// port is valid, good to go
		if app.Name != "" {
			apps[name] = app
		}
	}

	return apps
}

func (m MesosDiscoverer) tasks() ([]refinedMesosTask, error) {
	// TODO: ask some master, check if it is leading
	// TODO: if it is not leading master, go to leading
	// TODO: fail if leading master is actually empty
	resp, err := http.Get(m.masters[0] + "/state.json")
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	s := mesosState{}
	err = json.NewDecoder(resp.Body).Decode(&s)
	if err != nil {
		return nil, err
	}

	hosts := map[string]string{}
	for _, s := range s.Slaves {
		hosts[s.ID] = s.Host
	}

	tasks := []refinedMesosTask{}
	for _, f := range s.Frameworks {
		for _, t := range f.Tasks {
			if t.State != "TASK_RUNNING" {
				continue
			}

			if len(t.Resources.Ports) == 0 {
				continue
			}

			labels := map[string]string{}
			for _, l := range t.Labels {
				labels[l.Key] = l.Value
			}

			tasks = append(tasks, refinedMesosTask{
				Name:   t.Name,
				Host:   hosts[t.SlaveID],
				Ports:  t.Resources.Ports,
				Labels: labels,
			})
		}
	}

	return tasks, nil
}
