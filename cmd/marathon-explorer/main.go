package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bobrik/zoidberg"
	"github.com/gambol99/go-marathon"
	"github.com/samuel/go-zookeeper/zk"
)

func main() {
	m := flag.String("marathon", "", "marathon url")
	b := flag.String("balancer", "", "balancer app id")
	g := flag.String("groups", "", "comma-separated groups that hold app versions")
	z := flag.String("zk", "", "zk connection in host:port,host:port/path format")
	flag.Parse()

	if *m == "" || *b == "" || *g == "" || *z == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if os.Getenv("HOST") == "" || os.Getenv("PORT") == "" {
		log.Println("HOST and PORT environment variables should be set")
		os.Exit(1)
	}

	host := os.Getenv("HOST")
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		log.Fatal("error parsing port from environment:", err)
	}

	if !strings.Contains(*z, "/") {
		log.Fatal("zk connection string is invalid")
	}

	zz := strings.SplitN(*z, "/", 2)

	zh, zp := zz[0], "/"+zz[1]
	log.Println(zh, zp)

	zc, zch, err := zk.Connect(strings.Split(zh, ","), time.Minute)
	if err != nil {
		log.Fatal("error connecting to zookeeper:", err)
	}

	go func() {
		for e := range zch {
			log.Println("received zk event:", e)
		}
	}()

	mc, err := marathon.NewClient(marathon.Config{
		URL:            *m,
		RequestTimeout: 5,
	})
	if err != nil {
		log.Fatal(err)
	}

	l := zoidberg.ExplorerLocation{
		Host: host,
		Port: port,
	}

	d := zoidberg.NewMarathonDiscoverer(mc, *b, strings.Split(*g, ","))

	e, err := zoidberg.NewExplorer(d, zc, zp, l)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		addr := fmt.Sprintf("%s:%d", host, port)
		err := http.ListenAndServe(addr, e.ServeMux())
		if err != nil {
			log.Fatal(err)
		}
	}()

	err = e.Run()
	if err != nil {
		log.Fatal(err)
	}
}
