package balancer

import (
	"errors"
	"flag"
	"net"
	"os"
	"strconv"
	"strings"
)

var staticBalancersFlag *string

func init() {
	staticBalancersFlag = flag.String(
		"balancer-finder-static-balancers",
		os.Getenv("BALANCER_FINDER_STATIC_BALANCERS"),
		"list of balancers (host:port[,host:port]) for static balancer finder",
	)

	RegisterFinderMakerFromFlags("static", NewStaticFinderFromFlags)
}

// NewStaticFinderFromFlags returns new static finder from global flags
func NewStaticFinderFromFlags() (Finder, error) {
	if *staticBalancersFlag == "" {
		return nil, errors.New("got empty list of static balancers")
	}

	h := strings.Split(*staticBalancersFlag, ",")

	r := make([]Balancer, len(h))
	for i, hp := range h {
		b, err := balancerFromString(hp)
		if err != nil {
			return nil, err
		}

		r[i] = b
	}

	return NewStaticFinder(r), nil
}

// StaticFinder represents a finder that gets balancers from cli args
type StaticFinder struct {
	balancers []Balancer
}

// NewStaticFinder creates a new static Finder with
// the list of available load balancers
func NewStaticFinder(balancers []Balancer) StaticFinder {
	return StaticFinder{
		balancers: balancers,
	}
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
