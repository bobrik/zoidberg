package zoidberg

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/bobrik/zoidberg/application"
	"github.com/bobrik/zoidberg/balancer"
	"github.com/bobrik/zoidberg/state"
	"github.com/samuel/go-zookeeper/zk"
)

// Explorer constantly updates cluster state and notifies Balancers
type Explorer struct {
	name      string
	af        application.Finder
	bf        balancer.Finder
	zookeeper *zk.Conn
	zp        string
	state     state.State
	updated   map[string]update
	interval  time.Duration
	laziness  time.Duration
	mutex     sync.Mutex
}

// NewExplorer creates a new Explorer instance with a name,
// application and balancer finders and zookeeper connection
// to persist versioning information
func NewExplorer(name string, af application.Finder, bf balancer.Finder, zc *zk.Conn, zp string, interval, laziness time.Duration) (*Explorer, error) {
	s := state.State{}

	ss, _, err := zc.Get(zp)
	if err != nil {
		if err != zk.ErrNoNode {
			return nil, err
		}

		s.Versions = map[string]state.Versions{}
	} else {
		err := json.Unmarshal(ss, &s)
		if err != nil {
			return nil, err
		}
	}

	return &Explorer{
		name:      name,
		af:        af,
		bf:        bf,
		zookeeper: zc,
		zp:        zp,
		state:     s,
		updated:   map[string]update{},
		interval:  interval,
		laziness:  laziness,
		mutex:     sync.Mutex{},
	}, nil
}

// Run launches explorer's main loop that fetches state
// and updates load balancers' state
func (e *Explorer) Run() error {
	for {
		time.Sleep(e.interval)

		d, err := e.discover()
		if err != nil {
			log.Fatal("error discovering:", err)
			continue
		}

		e.updateBalancers(d)
	}
}

// discover returns the current view of the world
func (e *Explorer) discover() (*Discovery, error) {
	a, err := e.af.Apps()
	if err != nil {
		return nil, err
	}

	b, err := e.bf.Balancers()
	if err != nil {
		return nil, err
	}

	return &Discovery{
		Balancers: b,
		Apps:      a,
	}, nil
}

// updateBalancers updates state of all load balancers
// in parallel with the specified discovery information
func (e *Explorer) updateBalancers(discovery *Discovery) {
	state := e.getState()

	now := time.Now()

	updates := []balancer.Balancer{}
	for _, b := range discovery.Balancers {
		bs := b.String()
		if reflect.DeepEqual(e.updated[bs].apps, discovery.Apps) {
			if now.Sub(e.updated[bs].time) < e.laziness {
				continue
			}
		}

		updates = append(updates, b)
	}

	wg := sync.WaitGroup{}
	wg.Add(len(updates))

	for _, b := range updates {
		go func(b balancer.Balancer) {
			defer wg.Done()

			err := b.Update(e.name, discovery.Apps, state)
			if err != nil {
				log.Printf("error updating state on %s: %s\n", b, err)
				return
			}

			e.mutex.Lock()
			e.updated[b.String()] = update{
				time: now,
				apps: discovery.Apps,
			}
			e.mutex.Unlock()
		}(b)
	}

	wg.Wait()
}

// getState returns the current state of the world
func (e *Explorer) getState() state.State {
	e.mutex.Lock()
	s := e.state
	e.mutex.Unlock()

	return s
}

// setVersions sets version information for the specified application
func (e *Explorer) setVersions(app string, versions state.Versions) {
	e.mutex.Lock()
	e.state.Versions[app] = versions
	e.mutex.Unlock()
}

// persistState persists version state in zookeeper
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

// setUpZkPath initializes zookeeper if needed
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

// ServeMux returns explorer's api server
func (e *Explorer) ServeMux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/_health", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	mux.HandleFunc("/state", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "GET" {
			http.Error(w, "expected GET", http.StatusBadRequest)
			return
		}

		w.Header().Add("Content-type", "application/json")
		err := json.NewEncoder(w).Encode(e.getState())
		if err != nil {
			log.Println("error sending state:", err)
		}
	})

	mux.HandleFunc("/versions/", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "POST" && req.Method != "PUT" {
			http.Error(w, "expected POST or PUT", http.StatusBadRequest)
			return
		}

		d := json.NewDecoder(req.Body)
		v := state.Versions{}
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
		d, err := e.discover()
		if err != nil {
			http.Error(w, fmt.Sprintf("error getting servers: %s", err), http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-type", "application/json")
		err = json.NewEncoder(w).Encode(d)
		if err != nil {
			log.Println("error sending discovery:", err)
		}
	})

	return mux
}

type update struct {
	time time.Time
	apps application.Apps
}
