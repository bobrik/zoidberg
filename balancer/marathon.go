package balancer

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
	RegisterFinderMaker("marathon", FinderMaker{
		Flags: func() {
			marathonURLFlag = flag.String(
				"balancer-finder-marathon-url",
				os.Getenv("BALANCER_FINDER_MARATHON_URL"),
				"marathon url (http://host:port[,host:port]) for marathon balancer finder",
			)

			marathonFinderBalancer = flag.String(
				"balancer-finder-marathon-name",
				os.Getenv("BALANCER_FINDER_MARATHON_NAME"),
				"balancer name for marathon balancer finder",
			)
		},
		Maker: func() (Finder, error) {
			return NewMarathonFinder(*marathonURLFlag, *marathonFinderBalancer)
		},
	})
}

// MarathonFinder represents a finder that finds balancers in Marathon
type MarathonFinder struct {
	f *marathon.AppFetcher
	b string
}

// NewMarathonFinder creates a new Marathon Finder with
// Marathon location and load balancer name
func NewMarathonFinder(u string, b string) (*MarathonFinder, error) {
	if len(u) == 0 {
		return nil, errors.New("empty marathon url for marathon balancer finder")
	}

	if len(b) == 0 {
		return nil, errors.New("empty balancer name for marathon balancer finder")
	}

	f, err := marathon.NewAppFetcher(u)
	if err != nil {
		return nil, err
	}

	return &MarathonFinder{
		f: f,
		b: b,
	}, nil
}

// Balancers returns our load balancers running on associated Marathon
func (m *MarathonFinder) Balancers() ([]Balancer, error) {
	apps, err := m.f.FetchApps("zoidberg_balancer_for", m.b)
	if err != nil {
		return nil, err
	}

	balancers := []Balancer{}
	for _, app := range apps {
		if len(app.Ports) == 0 {
			log.Printf("app %s has no ports\n", app.ID)
			continue
		}

		for _, task := range app.Tasks {
			balancers = append(balancers, Balancer{
				Host: task.Host,
				Port: task.Ports[0],
			})
		}
	}

	return balancers, nil
}
