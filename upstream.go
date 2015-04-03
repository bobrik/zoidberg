package explorer

type Server struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

type Upstream struct {
	Host   string `json:"host"`
	Port   int    `json:"port"`
	Weight int    `json:"weight"`
}
