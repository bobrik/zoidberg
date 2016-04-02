package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bobrik/zoidberg"
	"github.com/bobrik/zoidberg/application"
	"github.com/bobrik/zoidberg/balancer"
	"github.com/samuel/go-zookeeper/zk"
)

func main() {
	n := flag.String("name", os.Getenv("NAME"), "zoidberg name")
	h := flag.String("host", os.Getenv("HOST"), "host")
	p := flag.String("port", os.Getenv("PORT"), "port")
	bff := flag.String("balancer-finder", os.Getenv("BALANCER_FINDER"), "balancer finder")
	aff := flag.String("application-finder", os.Getenv("APPLICATION_FINDER"), "application finder")
	z := flag.String("zk", os.Getenv("ZK"), "zk connection in host:port,host:port/path format")
	i := flag.Duration("interval", time.Second, "discovery interval")
	l := flag.Duration("laziness", time.Minute, "time to skip balancer updates if there are no changes")

	application.RegisterFlags()
	balancer.RegisterFlags()

	flag.Parse()

	if *bff == "" || *aff == "" || *h == "" || *p == "" || *z == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	bf, err := balancer.FinderByName(*bff)
	if err != nil {
		log.Fatal(err)
	}

	af, err := application.FinderByName(*aff)
	if err != nil {
		log.Fatal(err)
	}

	zc, zp, err := initZK(*z)
	if err != nil {
		log.Fatal(err)
	}

	e, err := zoidberg.NewExplorer(*n, af, bf, zc, zp, *i, *l)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		addr := fmt.Sprintf("%s:%s", *h, *p)
		log.Fatal(http.ListenAndServe(addr, e.ServeMux()))
	}()

	err = e.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func initZK(z string) (*zk.Conn, string, error) {
	if !strings.Contains(z, "/") {
		return nil, "", errors.New("zk connection string is invalid")
	}

	zz := strings.SplitN(z, "/", 2)

	zh, zp := zz[0], "/"+zz[1]

	zc, zch, err := zk.Connect(strings.Split(zh, ","), time.Minute)
	if err != nil {
		return nil, "", err
	}

	go func() {
		for e := range zch {
			log.Println("received zk event:", e)
		}
	}()

	return zc, zp, nil
}
