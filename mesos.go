package zoidberg

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

var ErrNoMesosMaster = errors.New("mesos master not found")

type refinedMesosTask struct {
	Name   string
	Host   string
	Ports  []int
	Labels map[string]string
}

type mesosState struct {
	Frameworks []mesosFramework `json:"frameworks"`
	Slaves     []mesosSlave     `json:"slaves"`
	Pid        string           `json:"pid"`
	Leader     string           `json:"leader"`
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
			log.Printf("task %s has no label zoidberg_app_version\n", task.Name)
			continue
		}

		version := task.Labels["zoidberg_app_version"]
		if version == "" {
			log.Printf("task %s has no label zoidberg_app_version\n", task.Name)
			continue
		}

		app := apps[name]
		if app.Name == "" {
			app.Name = name
			app.Servers = []Server{}
		}

		app.Servers = append(app.Servers, Server{
			Version: version,
			Host:    task.Host,
			Port:    task.Ports[0],
			Ports:   task.Ports,
		})

		apps[name] = app
	}

	return apps
}

func (m MesosDiscoverer) tasks() ([]refinedMesosTask, error) {
	s := mesosState{}

	for _, master := range m.masters {
		resp, err := http.Get(master + "/state.json")
		if err != nil {
			log.Printf("error fetching state from %s: %s\n", master, err)
			continue
		}

		defer resp.Body.Close()

		err = json.NewDecoder(resp.Body).Decode(&s)
		if err != nil {
			log.Printf("error decoding state from %s: %s\n", master, err)
			continue
		}

		if s.Pid != s.Leader {
			continue
		}

		return m.tasksFromLeader(s)
	}

	return nil, ErrNoMesosMaster
}

func (m MesosDiscoverer) tasksFromLeader(s mesosState) ([]refinedMesosTask, error) {
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
