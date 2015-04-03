package explorer

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/samuel/go-zookeeper/zk"
)

type Explorer struct {
	discoverer Discoverer
	zookeeper  *zk.Conn
	zp         string
	state      State
	location   ExplorerLocation
	err        error
	mutex      sync.Mutex
}

type ExplorerLocation struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

func NewExplorer(d Discoverer, zc *zk.Conn, zp string, location ExplorerLocation) (*Explorer, error) {
	state := State{}

	ss, _, err := zc.Get(zp)
	if err != nil {
		if err != zk.ErrNoNode {
			return nil, err
		}

		state.Versions = map[string]Version{}
	} else {
		err := json.Unmarshal(ss, &state)
		if err != nil {
			return nil, err
		}
	}

	return &Explorer{
		discoverer: d,
		zookeeper:  zc,
		zp:         zp,
		state:      state,
		location:   location,
		mutex:      sync.Mutex{},
	}, nil
}

func (e *Explorer) Run() error {
	for {
		time.Sleep(time.Second)

		d, err := e.discoverer.Discover()
		if err != nil {
			log.Println("error discovering:", err)
			continue
		}

		e.announce(d.Balancers)

		upstreams := e.state.GenerateUpstreams(d.Servers)
		e.updateUpstreams(d.Balancers, upstreams)

		sj, _ := json.Marshal(d.Servers)
		uj, _ := json.Marshal(upstreams)

		log.Println("servers:", string(sj))
		log.Println("upstreams:", string(uj))
	}

	return nil
}

func (e *Explorer) setError(err error) {
	e.mutex.Lock()
	e.err = err
	e.mutex.Unlock()
}

func (e *Explorer) announce(balancers []Balancer) {
	for _, b := range balancers {
		err := b.announce(e.location)
		if err != nil {
			log.Printf("error announcing itself to %s: %s\n", b, err)
			continue
		}

		log.Println("announced itself to", b)
	}
}

func (e *Explorer) updateUpstreams(balancers []Balancer, upstreams []Upstream) {
	for _, b := range balancers {
		err := b.updateUpstreams(upstreams)
		if err != nil {
			log.Printf("error updating upstreams on %s: %s\n", b, err)
			continue
		}

		log.Println("updated upstreams on", b)
	}
}

func (e *Explorer) getState() State {
	e.mutex.Lock()
	s := e.state
	e.mutex.Unlock()

	return s
}

func (e *Explorer) setState(s State) {
	e.mutex.Lock()
	e.state = s
	e.mutex.Unlock()
}

func (e *Explorer) persistState() error {
	s := e.getState()
	b, err := json.Marshal(s)
	if err != nil {
		return err
	}

	_, err = e.zookeeper.Set(e.zp, b, -1)
	if err == zk.ErrNoNode {
		_, err = e.zookeeper.Create(e.zp, b, 0, zk.WorldACL(zk.PermAll))
	}

	return err
}

func (e *Explorer) ServeMux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/_health", func(w http.ResponseWriter, req *http.Request) {
		e.mutex.Lock()
		defer e.mutex.Unlock()
		if e.err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	})

	mux.HandleFunc("/state", func(w http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()

		switch req.Method {
		case "GET":
			w.Header().Add("Content-type", "application/json")
			json.NewEncoder(w).Encode(e.getState())
		case "POST", "PUT":
			d := json.NewDecoder(req.Body)
			s := State{}
			err := d.Decode(&s)
			if err != nil {
				http.Error(w, fmt.Sprintf("state decoding failed: %s", err), http.StatusBadRequest)
				return
			}

			e.setState(s)
			err = e.persistState()
			if err != nil {
				http.Error(w, fmt.Sprintf("state set successfully, but persisting failed: %s", err), http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	})

	mux.HandleFunc("/discovery", func(w http.ResponseWriter, req *http.Request) {
		d, err := e.discoverer.Discover()
		if err != nil {
			http.Error(w, fmt.Sprintf("error getting servers: %s", err), http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-type", "application/json")
		json.NewEncoder(w).Encode(d)
	})

	return mux
}
