package explorer

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
	masters []string
	app     string
}

func NewMesosDiscoverer(masters []string, app string) MesosDiscoverer {
	return MesosDiscoverer{
		masters: masters,
		app:     app,
	}
}

func (m MesosDiscoverer) Discover() (Discovery, error) {
	discovery := Discovery{}

	tasks, err := m.tasks()
	if err != nil {
		return discovery, err
	}

	discovery.Balancers = m.balancers(tasks)
	discovery.Servers = m.servers(tasks)

	return discovery, nil
}

func (m MesosDiscoverer) balancers(tasks []refinedMesosTask) []Balancer {
	balancers := []Balancer{}

	for _, task := range tasks {
		if task.Labels["explorer_balancer_for"] == m.app {
			balancers = append(balancers, Balancer{
				Host: task.Host,
				Port: task.Ports[0],
			})
		}
	}

	return balancers
}

func (m MesosDiscoverer) servers(tasks []refinedMesosTask) map[string][]Server {
	servers := map[string][]Server{}

	for _, task := range tasks {
		if task.Labels["explorer_app"] == m.app {
			version := task.Labels["explorer_app_version"]

			servers[version] = append(servers[version], Server{
				Host: task.Host,
				Port: task.Ports[0],
			})
		}
	}

	return servers
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
