package application

import (
	"errors"
	"flag"
	"log"
	"os"

	"github.com/bobrik/zoidberg/marathon"
)

var marathonURLFlag *string
var marathonFinderBalancer *string

func init() {
	marathonURLFlag = flag.String(
		"application-finder-marathon-url",
		os.Getenv("APPLICATION_FINDER_MARATHON_URL"),
		"marathon url (http://host:port[,host:port]) for marathon application finder",
	)

	marathonFinderBalancer = flag.String(
		"application-finder-marathon-balancer",
		os.Getenv("APPLICATION_FINDER_MARATHON_BALANCER"),
		"balancer name for marathon application finder",
	)

	RegisterFinderMakerFromFlags("marathon", NewMarathonFinderFromFlags)
}

// NewMarathonFinderFromFlags returns new Marathon finder from global flags
func NewMarathonFinderFromFlags() (Finder, error) {
	return NewMarathonFinder(*marathonURLFlag, *marathonFinderBalancer)
}

// MarathonFinder represents a finder that finds apps in Marathon
type MarathonFinder struct {
	f        *marathon.AppFetcher
	balancer string
}

// NewMarathonFinder creates a new Marathon Finder with
// Marathon location and load balancer name
func NewMarathonFinder(u string, b string) (Finder, error) {
	if len(u) == 0 {
		return nil, errors.New("empty marathon url for marathon application finder")
	}

	if len(b) == 0 {
		return nil, errors.New("empty balancer name for marathon application finder")
	}

	f, err := marathon.NewAppFetcher(u)
	if err != nil {
		return nil, err
	}

	return &MarathonFinder{
		f:        f,
		balancer: b,
	}, nil
}

// Apps returns our applications running on associated Marathon
func (m *MarathonFinder) Apps() (Apps, error) {
	ma, err := m.f.Apps()
	if err != nil {
		return nil, err
	}

	apps := map[string]App{}

	for _, a := range ma {
		meta := parseLabels(a.Labels)
		for port, labels := range meta {
			if labels["balanced_by"] != m.balancer {
				continue
			}
			name := labels["app_name"]
			if name == "" {
				log.Printf("app %s has no label zoidberg_app_name\n", a.ID)
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

				// labels only come from the first task,
				// this could lead to funny errors if there
				// are multiple tasks with the same zoidberg_app_name
				app.Meta = labels
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
					Port:    task.Ports[port],
					Ports:   task.Ports,
				})
			}

			apps[name] = app
		}
	}

	return apps, nil
}
