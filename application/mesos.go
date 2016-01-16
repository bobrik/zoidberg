package application

import (
	"errors"
	"flag"
	"log"
	"os"
	"strings"

	"github.com/bobrik/zoidberg/mesos"
)

var mesosMastersFlag *string
var mesosFinderBalancerFlag *string

func init() {
	mesosMastersFlag = flag.String(
		"application-finder-mesos-masters",
		os.Getenv("APPLICATION_FINDER_MESOS_MASTERS"),
		"mesos masters (http://host:port[,http://host:port]) for mesos application finder",
	)

	mesosFinderBalancerFlag = flag.String(
		"application-finder-mesos-name",
		os.Getenv("APPLICATION_FINDER_MESOS_BALANCER"),
		"balancer name for mesos application finder",
	)

	RegisterFinderMakerFromFlags("mesos", NewMesosFinderFromFlags)
}

// NewMesosFinderFromFlags returns new Mesos finder from global flags
func NewMesosFinderFromFlags() (Finder, error) {
	return NewMesosFinder(strings.Split(*mesosMastersFlag, ","), *mesosFinderBalancerFlag)
}

// MesosFinder represents a finder that finds apps on Mesos
type MesosFinder struct {
	balancer string
	fetcher  *mesos.TaskFetcher
}

// NewMesosFinder creates a new Mesos Finder with
// Mesos master locations and load balancer name
func NewMesosFinder(masters []string, balancer string) (*MesosFinder, error) {
	if len(masters) == 0 {
		return nil, errors.New("empty list of masters for mesos balancer finder")
	}

	if len(balancer) == 0 {
		return nil, errors.New("empty balancer name for mesos balancer finder")
	}

	return &MesosFinder{
		balancer: balancer,
		fetcher:  mesos.NewTaskFetcher(masters),
	}, nil
}

// Apps returns our applications running on associated Mesos cluster
func (m *MesosFinder) Apps() (Apps, error) {
	tasks, err := m.fetcher.FetchTasks()
	if err != nil {
		return nil, err
	}

	apps := map[string]App{}

	for _, task := range tasks {
		meta := parseLabels(task.Labels)

		for port, tags := range meta {
			if task.Labels["balanced_by"] != m.balancer {
				continue
			}

			name := task.Labels["app_name"]
			if name == "" {
				log.Printf("task %s has no label app_name\n", task.Name)
				continue
			}

			version := task.Labels["app_version"]
			if version == "" {
				version = "1"
			}

			app := apps[name]
			if app.Name == "" {
				app.Name = name
				app.Servers = []Server{}

				// labels only come from the first task,
				// this could lead to funny errors if there
				// are multiple tasks with the same zoidberg_app_name
				app.Meta = tags
			}

			app.Servers = append(app.Servers, Server{
				Version: version,
				Host:    task.Host,
				Port:    task.Ports[port],
				Ports:   task.Ports,
			})

			apps[name] = app
		}
	}

	return apps, nil
}
