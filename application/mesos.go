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

func init() {
	RegisterFinderMaker("mesos", FinderMaker{
		Flags: func() {
			mesosMastersFlag = flag.String(
				"application-finder-mesos-masters",
				os.Getenv("APPLICATION_FINDER_MESOS_MASTERS"),
				"mesos masters (http://host:port[,http://host:port]) for mesos application finder",
			)
		},
		Maker: func(balancer string) (Finder, error) {
			return NewMesosFinder(strings.Split(*mesosMastersFlag, ","), balancer)
		},
	})
}

// MesosFinder represents a finder that finds apps on Mesos
type MesosFinder struct {
	fetcher  *mesos.TaskFetcher
	balancer string
}

// NewMesosFinder creates a new Mesos Finder with Mesos master locations
func NewMesosFinder(masters []string, balancer string) (*MesosFinder, error) {
	if len(masters) == 0 {
		return nil, errors.New("empty list of masters for mesos balancer finder")
	}

	return &MesosFinder{
		fetcher:  mesos.NewTaskFetcher(masters),
		balancer: balancer,
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
		for port, labels := range extractApps(task.Labels) {
			if m.balancer != labels["balanced_by"] {
				continue
			}

			name := labels["app_name"]
			if name == "" {
				log.Printf("task %s has no label zoidberg_port_%d_app_version", task.Name, port)
				continue
			}

			version := labels["app_version"]
			if version == "" {
				version = "1"
			}

			app := apps[name]
			if app.Name == "" {
				app.Name = name
				app.Servers = []Server{}
				app.Meta = labels
			}

			if port >= len(task.Ports) {
				log.Printf("task %s does not have expected port %d", task.Name, port)
				continue
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
