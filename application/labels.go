package application

import (
	"log"
	"strconv"
	"strings"
)

// parseLabels creates app labels from zoidberg prefixed labels
func parseLabels(labels map[string]string) map[int]map[string]string {
	r := make(map[int]map[string]string)
	for k, v := range labels {
		if strings.HasPrefix(k, "zoidberg_port_") {
			k2 := strings.TrimPrefix(k, "zoidberg_port_")
			k3 := strings.SplitN(k2, "_", 2)
			if p, err := strconv.Atoi(k3[0]); err != nil || len(k3) != 2 {
				log.Printf("found unparsable tag: %s", k)
				continue
			} else {
				if _, ok := r[p]; !ok {
					r[p] = make(map[string]string)
				}
				r[p][k3[1]] = v
			}
		}
	}
	return r
}
