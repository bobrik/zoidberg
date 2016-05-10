package application

import (
	"log"
	"strconv"
	"strings"
)

// extractApps returns map{ port => labels } for all found apps
func extractApps(labels map[string]string) map[int]map[string]string {
	r := map[int]map[string]string{}

	for k, v := range labels {
		if !strings.HasPrefix(k, "zoidberg_port_") {
			continue
		}

		p := strings.SplitN(strings.TrimPrefix(k, "zoidberg_port_"), "_", 2)
		if len(p) != 2 {
			log.Printf("invalid port: %s", k)
			continue
		}

		if port, err := strconv.Atoi(p[0]); err != nil {
			log.Printf("invalid port: %s", k)
			continue
		} else {
			if _, ok := r[port]; !ok {
				r[port] = map[string]string{}
			}
			r[port][p[1]] = v
		}
	}

	if _, ok := r[0]; !ok {
		a := extractLegacyApp(labels)
		if a != nil {
			r[0] = a
		}
	}

	return r
}

func extractLegacyApp(labels map[string]string) map[string]string {
	if labels["zoidberg_app_name"] == "" || labels["zoidberg_balanced_by"] == "" {
		// no legacy app present
		return nil
	}

	r := map[string]string{
		"app_name":    labels["zoidberg_app_name"],
		"balanced_by": labels["zoidberg_balanced_by"],
	}

	if labels["zoidberg_app_version"] != "" {
		r["app_version"] = labels["zoidberg_app_version"]
	}

	return r
}
