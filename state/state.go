package state

// State represents the state of apps
type State struct {
	Versions map[string]Versions `json:"versions"`
}

// Versions is a map of app names to their versions
type Versions map[string]Version

// Version represents some version and has a weight
type Version struct {
	// Weight is assigned directly to all tasks of the version
	Weight int `json:"weight"`
}
