package application

import (
	"fmt"
)

// Finder finds apps
type Finder interface {
	Apps() (Apps, error)
}

// FinderMakerFromFlags represents a function that makes
// a new Finder from global flags
type FinderMakerFromFlags func() (Finder, error)

// finderMakers contains a mapping of Finder names to their makers
var finderMakers = map[string]FinderMakerFromFlags{}

// RegisterFinderMakerFromFlags registers a new finder maker with a name
func RegisterFinderMakerFromFlags(name string, maker FinderMakerFromFlags) {
	finderMakers[name] = maker
}

// FinderByName returns existing finder by name
func FinderByName(finder string) (Finder, error) {
	if maker, ok := finderMakers[finder]; ok {
		return maker()
	}

	return nil, fmt.Errorf("unknown application finder: %q", finder)
}
