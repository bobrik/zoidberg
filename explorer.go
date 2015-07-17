package zoidberg

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/samuel/go-zookeeper/zk"
)

type Explorer struct {
	name       string
	discoverer Discoverer
	zookeeper  *zk.Conn
	zp         string
	state      State
	err        error
	mutex      sync.Mutex
}

func NewExplorer(name string, d Discoverer, zc *zk.Conn, zp string) (*Explorer, error) {
	state := State{}

	ss, _, err := zc.Get(zp)
	if err != nil {
		if err != zk.ErrNoNode {
			return nil, err
		}

		state.Versions = map[string]Versions{}
	} else {
		err := json.Unmarshal(ss, &state)
		if err != nil {
			return nil, err
		}
	}

	return &Explorer{
		name:       name,
		discoverer: d,
		zookeeper:  zc,
		zp:         zp,
		state:      state,
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

		e.updateBalancers(d)
	}
}

func (e *Explorer) setError(err error) {
	e.mutex.Lock()
	e.err = err
	e.mutex.Unlock()
}

func (e *Explorer) updateBalancers(discovery Discovery) {
	state := e.getState()

	for _, b := range discovery.Balancers {
		err := b.update(e.name, discovery.Apps, state)
		if err != nil {
			log.Printf("error updating state on %s: %s\n", b, err)
			continue
		}

		log.Println("updated state on", b)
	}
}

func (e *Explorer) getState() State {
	e.mutex.Lock()
	s := e.state
	e.mutex.Unlock()

	return s
}

func (e *Explorer) setVersions(app string, versions Versions) {
	e.mutex.Lock()
	e.state.Versions[app] = versions
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
		err = e.setUpZkPath(path.Dir(e.zp))
		if err != nil {
			return err
		}

		_, err = e.zookeeper.Create(e.zp, b, 0, zk.WorldACL(zk.PermAll))
	}

	return err
}

func (e *Explorer) setUpZkPath(p string) error {
	if p == "/" {
		return nil
	}

	err := e.setUpZkPath(path.Dir(p))
	if err != nil {
		return nil
	}

	_, err = e.zookeeper.Create(p, []byte{}, 0, zk.WorldACL(zk.PermAll))
	if err == zk.ErrNodeExists {
		return nil
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
		if req.Method != "GET" {
			http.Error(w, "expected GET", http.StatusBadRequest)
			return
		}

		w.Header().Add("Content-type", "application/json")
		json.NewEncoder(w).Encode(e.getState())
	})

	mux.HandleFunc("/versions/", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "POST" && req.Method != "PUT" {
			http.Error(w, "expected POST or PUT", http.StatusBadRequest)
			return
		}

		d := json.NewDecoder(req.Body)
		v := Versions{}
		err := d.Decode(&v)
		if err != nil {
			http.Error(w, fmt.Sprintf("version decoding failed: %s", err), http.StatusBadRequest)
			return
		}

		a := strings.TrimPrefix(req.URL.Path, "/versions/")
		if a == "" {
			http.Error(w, "application is not specified", http.StatusBadRequest)
			return
		}

		e.setVersions(a, v)
		err = e.persistState()
		if err != nil {
			http.Error(w, fmt.Sprintf("state set successfully, but persisting failed: %s", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
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
