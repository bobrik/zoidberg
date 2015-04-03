package explorer

type Discoverer interface {
	Discover() (Discovery, error)
}

type Discovery struct {
	Balancers []Balancer          `json:"balancers"`
	Servers   map[string][]Server `json:"servers"`
}
