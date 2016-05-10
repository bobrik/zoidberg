package application

import (
	"errors"
	"flag"
	"log"
	"os"

	fetcher "github.com/bobrik/zoidberg/marathon"
	"github.com/gambol99/go-marathon"
)

var marathonURLFlag *string

func init() {
	RegisterFinderMaker("marathon", FinderMaker{
		Flags: func() {
			marathonURLFlag = flag.String(
				"application-finder-marathon-url",
				os.Getenv("APPLICATION_FINDER_MARATHON_URL"),
				"marathon url (http://host:port[,host:port]) for marathon application finder",
			)
		},
		Maker: func(balancer string) (Finder, error) {
			return NewMarathonFinder(*marathonURLFlag, balancer)
		},
	})
}

// MarathonFinder represents a finder that finds apps in Marathon
type MarathonFinder struct {
	fetcher  *fetcher.AppFetcher
	balancer string
}

// NewMarathonFinder creates a new Marathon Finder with Marathon location
func NewMarathonFinder(url string, balancer string) (Finder, error) {
	if len(url) == 0 {
		return nil, errors.New("empty marathon url for marathon application finder")
	}

	fetcher, err := fetcher.NewAppFetcher(url)
	if err != nil {
		return nil, err
	}

	return &MarathonFinder{
		fetcher:  fetcher,
		balancer: balancer,
	}, nil
}

// Apps returns our applications running on associated Marathon
func (m *MarathonFinder) Apps() (Apps, error) {
	ma, err := m.fetcher.FetchApps(nil)
	if err != nil {
		return nil, err
	}

	apps := map[string]App{}

	for _, a := range ma {
		for port, labels := range extractApps(a.Labels) {
			if m.balancer != labels["balanced_by"] {
				continue
			}

			name := labels["app_name"]
			if name == "" {
				log.Printf("app %s has no label zoidberg_port_%d_app_name", a.ID, port)
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

			for _, task := range a.Tasks {
				server := marathonTaskToServer(task, port, version)
				if server != nil {
					app.Servers = append(app.Servers, *server)
				}
			}

			apps[name] = app
		}
	}

	return apps, nil
}

func marathonTaskToServer(task *marathon.Task, port int, version string) *Server {
	if port >= len(task.Ports) {
		log.Printf("task %s does not have expected port %d", task.ID, port)
		return nil
	}

	healthy := true
	for _, check := range task.HealthCheckResults {
		if check == nil {
			continue
		}

		if !check.Alive {
			healthy = false
			break
		}
	}

	if !healthy {
		return nil
	}

	return &Server{
		Version: version,
		Host:    task.Host,
		Port:    task.Ports[port],
		Ports:   task.Ports,
	}
}
