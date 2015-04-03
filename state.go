package explorer

type State struct {
	Versions map[string]Version `json:"versions"`
}

type Version struct {
	Name string `json:"name"`
	// Weight is assigned directly to all tasks of the version
	Weight int `json:"weight"`
}

func NewState(versions map[string]Version) State {
	return State{versions}
}

func (s State) GenerateUpstreams(servers map[string][]Server) []Upstream {
	upstreams := []Upstream{}

	for _, v := range s.Versions {
		if v.Weight == 0 {
			continue
		}

		if len(servers[v.Name]) == 0 {
			continue
		}

		for _, s := range servers[v.Name] {
			upstreams = append(upstreams, Upstream{
				Host:   s.Host,
				Port:   s.Port,
				Weight: v.Weight,
			})
		}
	}

	return upstreams
}
