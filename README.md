# Zoidberg

Zoidberg provides per-app service discovery for mesos with pluggable
discovery mechanisms. It allows you to shift traffic from version
to version in small percentage to ensure smooth deployments. It also
allows usual service discovery where it's up to framework how to
schedule tasks to avoid downtime.

## Compatible load balancers

* [zoidberg-nginx](https://github.com/bobrik/zoidberg-nginx)
supports upstream list updates without spawning new workers,
providing rich module ecosystem from nginx, including lua modules.

## Stability

Zoidberg is still experimental software. Don't expect it to work after
you blindly upgrade it. Better stick to one version and upgrade carefully.

## Architecture

Zoidberg consists of several parts:

* Discoverer is responsible for discovering cluster state.
* Load balancers are responsible for providing well known endpoints.
* Explorer is bounded with discoverer and responsible for version management.

### Discoverer

Discoverer is a mechanism to discover load balancers and application tasks.

#### Label-based discoverers

Both marathon and mesos discoverers are label base discoverers, which means
that they rely on labels to discover current state.

Used labels for apps:

* `zoidberg_app_name` defines application name.
* `zoidberg_app_version` defines application version.
* `zoidberg_app_port` defines main port for application.
* `zoidberg_balanced_by` defines load balancer name for application.

Used labels for balancers:

* `zoidberg_balancer_for` defines load balancer name.

Several applications can be published to a single load balancer.

#### Marathon discoverer

Marathon discoverer provides service discovery for Marathon.
It reads cluster state from Marathon API, providing health check
awareness, but it is also can be a bit slower with many Zoidberg instances.

Running:

```
docker run -rm -it -e HOST=127.0.0.1 -e PORT=12345 bobrik/zoidberg:0.3.0 \
    /go/bin/marathon-explorer -balancer mybalancer \
    -marathon http://marathon.dev:8080 -zk zk:2181/zoidberg-marathon-mybalancer
```

For setup with several Marathon nodes you can use the following syntax:

```
http://marathon1:8080,marathon2:8080,marathon3:8080
```

#### Mesos discoverer

Mesos discoverer provides service discovery for any framework running on mesos.
It reads cluster state from Mesos master and is unaware of health checks
that can be defined in framework. It also doesn't support several masters.

Running:

```
docker run --rm -it -e HOST=127.0.0.1 -e PORT=12345 bobrik/zoidberg:0.3.0 \
    /go/bin/mesos-explorer -balancer mybalancer \
    -master http://mesos-master:5050 -zk zk:2181/zoidberg-mesos-mybalancer
```

### Load balancers

Load balancers are responsible for providing service endpoints at well-known
address and load balancing. Zoidberg requires balancers to implement
the next HTTP API:

* `PUT /state` or `POST /state` with json like this:

```json
{
  "apps": {
    "myapp": {
      "name": "myapp",
      "port": 13131,
      "servers": [
        {
          "host": "192.168.0.7",
          "port": 31633,
          "version": "/v1"
        }
      ]
    }
  },
  "state": {
    "versions": {
      "myapp": {
        "/v1": {
          "name": "/v1",
          "weight": 2
        }
      }
    }
  },
  "explorer": {
    "host": "127.0.0.1",
    "port": 38999
  }
}
```

Apps have associated versions and it's generally up to balancer to decide
how to use them. General approach is to use provided weights for servers
of specified versions.

Explorer also exposes own location to use.

### Explorer

Explorer is responsible for distributing state of Mesos cluster to balancers.
It is also responsible for switching traffic between application versions.

Explorer provides the next HTTP API:

* `PUT /versions/{{app}}` or `POST /versions/{{app}}` with json like this:

```json
{
  "/v1": {
    "name": "/v1",
    "weight": 2
  }
}
```

`{{app}}` in URL should be replaced with the name of an actual app.

* `GET /state` that returns full state

* `GET /discovery` that returns json like this:

```json
{
  "balancers": [
    {
      "host": "192.168.0.7",
      "port": 31631
    }
  ],
  "apps": {
    "myapp": {
      "name": "myapp",
      "port": 13131,
      "servers": [
        {
          "host": "192.168.0.7",
          "port": 31000,
          "ports": [31000],
          "version": "/v1"
        }
      ]
    }
  }
}
```

* `GET /_health` that returns 2xx code if everything looks good

## Why?

![zoidberg](zoidberg.jpg)
