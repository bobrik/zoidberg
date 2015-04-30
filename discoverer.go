package zoidberg

type Discoverer interface {
	Discover() (Discovery, error)
}

type Discovery struct {
	Balancers []Balancer `json:"balancers"`
	Apps      Apps       `json:"apps"`
}
