package balancer

import (
	"errors"
	"flag"
	"os"
	"strings"

	"github.com/bobrik/zoidberg/mesos"
)

var mesosFinderMesosMastersFlag *string

func init() {
	RegisterFinderMaker("mesos", FinderMaker{
		Flags: func() {
			mesosFinderMesosMastersFlag = flag.String(
				"balancer-finder-mesos-masters",
				os.Getenv("BALANCER_FINDER_MESOS_MASTERS"),
				"mesos masters (http://host:port[,http://host:port]) for mesos balancer finder",
			)
		},
		Maker: func(balancer string) (Finder, error) {
			return NewMesosFinder(strings.Split(*mesosFinderMesosMastersFlag, ","), balancer)
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

// Name returns the name of the balancer group
func (m *MesosFinder) Name() string {
	return m.balancer
}

// Balancers returns our load balancers running on Mesos
func (m *MesosFinder) Balancers() ([]Balancer, error) {
	tasks, err := m.fetcher.FetchTasks()
	if err != nil {
		return nil, err
	}

	balancers := []Balancer{}

	for _, task := range tasks {
		if task.Labels["zoidberg_balancer_for"] == m.balancer {
			balancers = append(balancers, Balancer{
				Host: task.Host,
				Port: task.Ports[0],
			})
		}
	}

	return balancers, nil
}
