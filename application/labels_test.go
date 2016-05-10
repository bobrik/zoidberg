package application

import (
	"reflect"
	"testing"
)

func TestLabels(t *testing.T) {
	table := []struct {
		labels map[string]string
		apps   map[int]map[string]string
	}{
		{
			labels: map[string]string{},
			apps:   map[int]map[string]string{},
		},
		{
			labels: map[string]string{
				"zoidberg_port_0_app_name":    "foo",
				"zoidberg_port_0_balanced_by": "bar",
			},
			apps: map[int]map[string]string{
				0: {
					"app_name":    "foo",
					"balanced_by": "bar",
				},
			},
		},
		{
			labels: map[string]string{
				"zoidberg_port_0_app_name":    "foo",
				"zoidberg_port_0_app_version": "8",
				"zoidberg_port_0_balanced_by": "bar",
			},
			apps: map[int]map[string]string{
				0: {
					"app_name":    "foo",
					"app_version": "8",
					"balanced_by": "bar",
				},
			},
		},
		{
			labels: map[string]string{
				"zoidberg_port_1_app_name":    "foo",
				"zoidberg_port_1_balanced_by": "bar",
			},
			apps: map[int]map[string]string{
				1: {
					"app_name":    "foo",
					"balanced_by": "bar",
				},
			},
		},
		{
			labels: map[string]string{
				"zoidberg_port_0_app_name":    "foo1",
				"zoidberg_port_0_balanced_by": "bar1",
				"zoidberg_port_1_app_name":    "foo2",
				"zoidberg_port_1_balanced_by": "bar2",
			},
			apps: map[int]map[string]string{
				0: {
					"app_name":    "foo1",
					"balanced_by": "bar1",
				},
				1: {
					"app_name":    "foo2",
					"balanced_by": "bar2",
				},
			},
		},
		{
			labels: map[string]string{
				"zoidberg_app_name":    "foo",
				"zoidberg_balanced_by": "bar",
			},
			apps: map[int]map[string]string{
				0: {
					"app_name":    "foo",
					"balanced_by": "bar",
				},
			},
		},
		{
			labels: map[string]string{
				"zoidberg_app_name":    "foo",
				"zoidberg_app_version": "8",
				"zoidberg_balanced_by": "bar",
			},
			apps: map[int]map[string]string{
				0: {
					"app_name":    "foo",
					"app_version": "8",
					"balanced_by": "bar",
				},
			},
		},
		{
			labels: map[string]string{
				"zoidberg_port_0_app_name":    "foo1",
				"zoidberg_port_0_balanced_by": "bar1",
				"zoidberg_app_name":           "foo2",
				"zoidberg_balanced_by":        "bar2",
			},
			apps: map[int]map[string]string{
				0: {
					"app_name":    "foo1",
					"balanced_by": "bar1",
				},
			},
		},
	}

	for _, row := range table {
		r := extractApps(row.labels)
		if !reflect.DeepEqual(r, row.apps) {
			t.Errorf("expected: %v, got: %v", row.apps, r)
		}
	}
}
