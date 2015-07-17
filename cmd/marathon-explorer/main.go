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
	n := flag.String("name", "", "zoidberg name")
	m := flag.String("marathon", "", "marathon url")
	b := flag.String("balancer", "", "balancer name")
	s := flag.String("servers", "", "static list of servers to use as balancers")
	z := flag.String("zk", "", "zk connection in host:port,host:port/path format")
	flag.Parse()

	if *n == "" || *m == "" || *b == "" || *z == "" {
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

	d := zoidberg.NewMarathonDiscoverer(mc, *b)

	balancers, err := zoidberg.BalancersFromString(*s)
	if err != nil {
		log.Fatal(err)
	}

	if balancers != nil {
		d.SetStaticBalancers(balancers)
	}

	e, err := zoidberg.NewExplorer(*n, d, zc, zp)
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
