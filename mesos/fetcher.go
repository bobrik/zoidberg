package mesos

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"
)

// ErrNoMaster indicates that no alive mesos masters are found
var ErrNoMesosMaster = errors.New("mesos master not found")

// TaskFetcher fetches Mesos tasks from Mesos masters
type TaskFetcher struct {
	masters []string
	client  http.Client
}

// NewTaskFetcher creates a new TaskFetcher with specified Mesos masters
func NewTaskFetcher(masters []string) *TaskFetcher {
	return &TaskFetcher{
		masters: masters,
		client: http.Client{
			Timeout: time.Second * 5,
		},
	}
}

// FetchTasks returns tasks currently running on Mesos cluster
func (f *TaskFetcher) FetchTasks() ([]Task, error) {
	s := mesosState{}

	for _, master := range f.masters {
		u := master + "/state.json"
		resp, err := f.client.Get(u)
		if err != nil {
			log.Printf("error fetching state from %s: %s\n", u, err)
			continue
		}

		defer func() {
			err := resp.Body.Close()
			if err != nil {
				log.Printf("error closing body from %s: %s\n", u, err)
			}
		}()

		err = json.NewDecoder(resp.Body).Decode(&s)
		if err != nil {
			log.Printf("error decoding state from %s: %s\n", u, err)
			continue
		}

		if s.Pid != s.Leader {
			continue
		}

		return f.tasksFromLeader(s)
	}

	return nil, ErrNoMesosMaster
}

// tasksFromLeader returns tasks from the currently leading Mesos master
func (f *TaskFetcher) tasksFromLeader(s mesosState) ([]Task, error) {
	hosts := map[string]string{}
	for _, s := range s.Slaves {
		hosts[s.ID] = s.Host
	}

	tasks := []Task{}
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

			tasks = append(tasks, Task{
				Name:   t.Name,
				Host:   hosts[t.SlaveID],
				Ports:  t.Resources.Ports,
				Labels: labels,
			})
		}
	}

	return tasks, nil
}
