package balancer

import (
	"errors"
	"flag"
	"log"
	"os"

	"github.com/bobrik/zoidberg/marathon"
)

var marathonFinderMarathonURLFlag *string

func init() {
	RegisterFinderMaker("marathon", FinderMaker{
		Flags: func() {
			marathonFinderMarathonURLFlag = flag.String(
				"balancer-finder-marathon-url",
				os.Getenv("BALANCER_FINDER_MARATHON_URL"),
				"marathon url (http://host:port[,host:port]) for marathon balancer finder",
			)
		},
		Maker: func(balancer string) (Finder, error) {
			return NewMarathonFinder(*marathonFinderMarathonURLFlag, balancer)
		},
	})
}

// MarathonFinder represents a finder that finds balancers in Marathon
type MarathonFinder struct {
	fetcher  *marathon.AppFetcher
	balancer string
}

// NewMarathonFinder creates a new Marathon Finder with
// Marathon location and load balancer name
func NewMarathonFinder(url string, balancer string) (*MarathonFinder, error) {
	if len(url) == 0 {
		return nil, errors.New("empty marathon url for marathon balancer finder")
	}

	if len(balancer) == 0 {
		return nil, errors.New("empty balancer name for marathon balancer finder")
	}

	fetcher, err := marathon.NewAppFetcher(url)
	if err != nil {
		return nil, err
	}

	return &MarathonFinder{
		fetcher:  fetcher,
		balancer: balancer,
	}, nil
}

// Name returns the name of the balancer group
func (m *MarathonFinder) Name() string {
	return m.balancer
}

// Balancers returns our load balancers running on associated Marathon
func (m *MarathonFinder) Balancers() ([]Balancer, error) {
	apps, err := m.fetcher.FetchApps(map[string]string{"zoidberg_balancer_for": m.balancer})
	if err != nil {
		return nil, err
	}

	balancers := []Balancer{}
	for _, app := range apps {
		if len(app.Ports) == 0 {
			log.Printf("app %s has no ports", app.ID)
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
