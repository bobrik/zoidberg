package mesos

import (
	"fmt"
	"strconv"
	"strings"
)

type mesosState struct {
	Frameworks []mesosFramework `json:"frameworks"`
	Slaves     []mesosSlave     `json:"slaves"`
	Pid        string           `json:"pid"`
	Leader     string           `json:"leader"`
}

type mesosFramework struct {
	Tasks []mesosTask `json:"tasks"`
}

type mesosTask struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	State     string         `json:"state"`
	SlaveID   string         `json:"slave_id"`
	Resources mesosResources `json:"resources"`
	Labels    []mesosLabel   `json:"labels"`
}

type mesosSlave struct {
	ID   string `json:"id"`
	Host string `json:"hostname"`
}

type mesosResources struct {
	Ports mesosPorts `json:"ports"`
}

type mesosLabel struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type mesosPorts []int

func (mp *mesosPorts) UnmarshalJSON(b []byte) error {
	p := strings.TrimFunc(string(b), func(r rune) bool {
		return r == '[' || r == ']' || r == '"'
	})

	ports := []int{}

	for _, s := range strings.Split(p, " ") {
		r := strings.Split(strings.TrimSuffix(s, ","), "-")
		if len(r) != 2 {
			return fmt.Errorf("expected port range to be in formate XXXX-YYYY, got %s", r)
		}

		start, err := strconv.Atoi(r[0])
		if err != nil {
			return fmt.Errorf("error processing por range %q: %s", p, err)
		}

		end, err := strconv.Atoi(r[1])
		if err != nil {
			return fmt.Errorf("error processing por range %q: %s", p, err)
		}

		for i := start; i <= end; i++ {
			ports = append(ports, i)
		}
	}

	*mp = mesosPorts(ports)

	return nil
}
