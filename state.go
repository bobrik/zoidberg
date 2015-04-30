package zoidberg

type State struct {
	Versions map[string][]Version `json:"versions"`
}

type Version struct {
	Name string `json:"name"`
	// Weight is assigned directly to all tasks of the version
	Weight int `json:"weight"`
}
