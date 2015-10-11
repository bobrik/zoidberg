package application

// App represents a single application
type App struct {
	Name    string   `json:"name"`
	Servers []Server `json:"servers"`
}

// Apps is a map of app names to app instances
type Apps map[string]App

// Server is an instance of an app
type Server struct {
	Host    string `json:"host"`
	Port    int    `json:"port"`
	Ports   []int  `json:"ports"`
	Version string `json:"version"`
}
