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
	RegisterFinderMaker("mesos", FinderMaker{
		Flags: func() {
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
		},
		Maker: func() (Finder, error) {
			return NewMesosFinder(strings.Split(*mesosMastersFlag, ","), *mesosFinderBalancerFlag)
		},
	})
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
			version = "1"
		}

		app := apps[name]
		if app.Name == "" {
			app.Name = name
			app.Servers = []Server{}

			// labels only come from the first task,
			// this could lead to funny errors
			app.Meta = metaFromLabels(task.Labels)
		}

		for k, v := range task.Labels {
			if strings.HasPrefix(k, "zoidberg_meta_") {
				app.Meta[strings.TrimPrefix(k, "zoidberg_meta_")] = v
			}
		}

		app.Servers = append(app.Servers, Server{
			Version: version,
			Host:    task.Host,
			Port:    task.Ports[0],
			Ports:   task.Ports,
		})

		apps[name] = app
	}

	return apps, nil
}
