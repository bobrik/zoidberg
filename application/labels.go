package application

import "strings"

// metaFromLabels creates app labels from zoidberg prefixed labels
func metaFromLabels(labels map[string]string) map[string]string {
	r := map[string]string{}

	for k, v := range labels {
		if strings.HasPrefix(k, "zoidberg_meta_") {
			r[strings.TrimPrefix(k, "zoidberg_meta_")] = v
		}
	}

	return r
}
