package balancer

import (
	"fmt"
)

// Finder finds balancers
type Finder interface {
	Balancers() ([]Balancer, error)
}

// FinderMaker represents finder maker tuple:
// * a function to register flags
// * a function to make finder from parsed flags
type FinderMaker struct {
	Flags func()
	Maker func() (Finder, error)
}

// finderMakers contains a mapping of Finder names to their makers
var finderMakers = map[string]FinderMaker{}

// RegisterFinderMaker registers a new finder maker with a name
func RegisterFinderMaker(name string, maker FinderMaker) {
	finderMakers[name] = maker
}

// FinderByName returns existing finder by name
func FinderByName(finder string) (Finder, error) {
	if maker, ok := finderMakers[finder]; ok {
		return maker.Maker()
	}

	return nil, fmt.Errorf("unknown balancer finder: %q", finder)
}

// RegisterFlags registers flags of all finder makers
func RegisterFlags() {
	for _, m := range finderMakers {
		m.Flags()
	}
}
