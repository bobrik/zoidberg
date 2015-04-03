package explorer

import (
	"reflect"
	"testing"
)

type stateTest struct {
	state     State
	servers   map[string][]Server
	upstreams []Upstream
}

var stateTestTable = []stateTest{
	{
		state:     NewState(map[string]Version{}),
		servers:   map[string][]Server{},
		upstreams: []Upstream{},
	},
	{
		state: NewState(map[string]Version{}),
		servers: map[string][]Server{
			"/v1": []Server{
				{
					Host: "host1",
					Port: 12345,
				},
			},
		},
		upstreams: []Upstream{},
	},
	{
		state: NewState(map[string]Version{
			"/v1": Version{
				Name:   "/v1",
				Weight: 1,
			},
		}),
		servers: map[string][]Server{
			"/v2": []Server{
				{
					Host: "host1",
					Port: 12345,
				},
			},
		},
		upstreams: []Upstream{},
	},
	{
		state: NewState(map[string]Version{
			"/v1": Version{
				Name:   "/v1",
				Weight: 8,
			},
		}),
		servers: map[string][]Server{
			"/v1": []Server{
				{
					Host: "host1",
					Port: 12345,
				},
			},
		},
		upstreams: []Upstream{
			{
				Host:   "host1",
				Port:   12345,
				Weight: 8,
			},
		},
	},
	{
		state: NewState(map[string]Version{
			"/v1": Version{
				Name:   "/v1",
				Weight: 2,
			},
		}),
		servers: map[string][]Server{
			"/v1": []Server{
				{
					Host: "host1",
					Port: 12345,
				},
				{
					Host: "host2",
					Port: 12346,
				},
			},
		},
		upstreams: []Upstream{
			{
				Host:   "host1",
				Port:   12345,
				Weight: 2,
			},
			{
				Host:   "host2",
				Port:   12346,
				Weight: 2,
			},
		},
	},
	{
		state: NewState(map[string]Version{
			"/v1": Version{
				Name:   "/v1",
				Weight: 0,
			},
			"/v2": Version{
				Name:   "/v2",
				Weight: 2,
			},
			"/v3": Version{
				Name:   "/v3",
				Weight: 8,
			},
		}),
		servers: map[string][]Server{
			"/v1": []Server{
				{
					Host: "host1_1",
					Port: 12345,
				},
				{
					Host: "host1_2",
					Port: 12346,
				},
			},
			"/v2": []Server{
				{
					Host: "host2_1",
					Port: 22345,
				},
				{
					Host: "host2_2",
					Port: 22346,
				},
			},
			"/v3": []Server{
				{
					Host: "host3_1",
					Port: 32345,
				},
				{
					Host: "host3_2",
					Port: 32346,
				},
			},
		},
		upstreams: []Upstream{
			{
				Host:   "host2_1",
				Port:   22345,
				Weight: 2,
			},
			{
				Host:   "host2_2",
				Port:   22346,
				Weight: 2,
			},
			{
				Host:   "host3_1",
				Port:   32345,
				Weight: 8,
			},
			{
				Host:   "host3_2",
				Port:   32346,
				Weight: 8,
			},
		},
	},
}

func TestTableConversion(t *testing.T) {
	for _, test := range stateTestTable {
		upstreams := test.state.GenerateUpstreams(test.servers)

		if len(test.upstreams) != len(upstreams) {
			t.Errorf("expected %d upstreams, got %d", len(test.upstreams), len(upstreams))
		}

		if !reflect.DeepEqual(test.upstreams, upstreams) {
			t.Errorf("expected %#v, got %#v", test.upstreams, upstreams)
		}
	}
}
