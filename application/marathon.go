package application

import (
	"errors"
	"flag"
	"log"
	"os"

	"github.com/bobrik/zoidberg/marathon"
	m2 "github.com/gambol99/go-marathon"
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
	finder   *marathon.AppFetcher
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

	finder, err := marathon.NewAppFetcher(u)
	if err != nil {
		return nil, err
	}

	return &MarathonFinder{
		finder:   finder,
		balancer: b,
	}, nil
}

// Apps returns our applications running on associated Marathon
func (m *MarathonFinder) Apps() (Apps, error) {
	ma, err := m.finder.FetchApps(nil)
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
				if !healthCheck(task) {
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

func healthCheck(t *m2.Task) bool {
	for _, check := range t.HealthCheckResult {
		if check == nil {
			continue
		}

		if !check.Alive {
			return false
		}
	}
	return true
}
