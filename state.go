package zoidberg

type State struct {
	Versions map[string]Versions `json:"versions"`
}

type Versions map[string]Version

type Version struct {
	Name string `json:"name"`
	// Weight is assigned directly to all tasks of the version
	Weight int `json:"weight"`
}
