package zoidberg

type App struct {
	Name    string   `json:"name"`
	Port    int      `json:"port"`
	Servers []Server `json:"servers"`
}

type Apps map[string]App

type Server struct {
	Host    string `json:"host"`
	Port    int    `json:"port"`
	Ports   []int  `json:"ports"`
	Version string `json:"version"`
}
