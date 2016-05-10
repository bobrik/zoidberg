package balancer

import (
	"errors"
	"flag"
	"net"
	"os"
	"strconv"
	"strings"
)

var staticFinderBalancersFlag *string

func init() {
	RegisterFinderMaker("static", FinderMaker{
		Flags: func() {
			staticFinderBalancersFlag = flag.String(
				"balancer-finder-static-balancers",
				os.Getenv("BALANCER_FINDER_STATIC_BALANCERS"),
				"list of balancers (host:port[,host:port]) for static balancer finder",
			)
		},
		Maker: func(balancer string) (Finder, error) {
			if *staticFinderBalancersFlag == "" {
				return nil, errors.New("got empty list of static balancers")
			}

			h := strings.Split(*staticFinderBalancersFlag, ",")

			r := make([]Balancer, len(h))
			for i, hp := range h {
				b, err := balancerFromString(hp)
				if err != nil {
					return nil, err
				}

				r[i] = b
			}

			return NewStaticFinder(r, balancer), nil
		},
	})
}

// StaticFinder represents a finder that gets balancers from cli args
type StaticFinder struct {
	balancers []Balancer
	balancer  string
}

// NewStaticFinder creates a new static Finder with
// the list of available load balancers
func NewStaticFinder(balancers []Balancer, balancer string) StaticFinder {
	return StaticFinder{
		balancers: balancers,
		balancer:  balancer,
	}
}

// Name returns the name of the balancer group
func (s StaticFinder) Name() string {
	return s.balancer
}

// Balancers returns the static list of load balancers
func (s StaticFinder) Balancers() ([]Balancer, error) {
	return s.balancers, nil
}

// balanceFromString creates Balancer instance from a host:port string
func balancerFromString(s string) (Balancer, error) {
	b := Balancer{}

	h, p, err := net.SplitHostPort(s)
	if err != nil {
		return b, err
	}

	port, err := strconv.Atoi(p)
	if err != nil {
		return b, err
	}

	b.Host = h
	b.Port = port

	return b, nil
}
