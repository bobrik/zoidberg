package mesos

// Task represents a single running Mesos task
type Task struct {
	Name   string
	Host   string
	Ports  []int
	Labels map[string]string
}
