# Zoidberg

Zoidberg provides per-app service discovery for mesos with pluggable
discovery mechanisms. It allows you to shift traffic from version
to version in small percentage to ensure smooth deployments. It also
allows usual service discovery where it's up to framework how to
schedule tasks to avoid downtime.

## Architecture

Zoidberg consists of several parts:

### Discoverer

Discoverer is a mechanism to discover load balancers and application tasks.

#### Marathon discoverer

Marathon discoverer provides service discovery for marathon.

Balancer is provided as application id.

Applications are provided as their root groups. Application groups should only
hold application versions. Application name should be placed into
label `zoidberg_app_name` to make application discoverable.

Example:

* `/whatever/balancer` app holds tasks for balancer.
* `/whatever/app/v1` holds tasks for application version 1.

If you want to perform slow upgrade, you should deploy `/whatever/app/v2` and
shift traffic gradually with zoidberg API. If you are comfortable with rolling
upgrade using marathon, feel free to redeploy `v1`.

Running:

```
HOST=127.0.0.1 PORT=12345 marathon-explorer \
    -balancer /whatever/balancer \
    -groups /whatever/app \
    -marathon http://marathon:8080 \
    -zk zk:2181/zoidberg-whaterver
```

You can use comma-separated list of groups to discover multiple applications
with single load balancer.

#### Mesos discoverer

Mesos discoverer provides service discovery for any framework running on mesos.

Explorer is launched for specific named balancer.

Balancer is discovered by label `zoidberg_balancer_for`.

Applications are discovered by label `zoidberg_balanced_by`. Applications
should have labels `zoidberg_app_name`, `zoidberg_app_version` and
`zoidberg_app_port` to be discoverable.

It is possible to run tasks on marathon to make them discoverable
with both marathon and mesos discoverers.

Running:

```
HOST=127.0.0.1 PORT=12345 mesos-explorer \
 -balancer mybalancer \
 -master http://mesos-master:5050 \ 
 -zk zk:2181/zoidberg-mybalancer
```

### Load balancers

Load balancers are responsible for providing service endpoints
at well-known address and load balancing. Zoidberg requires balancers
to implement the next HTTP API:

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
      "myapp": [
        {
          "name": "/v1",
          "weight": 2
        }
      ]
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

Explorer is responsible for distributing state of mesos cluster to balancers.
It is also responsible for switching traffic between application versions.

Explorer provides the next HTTP API:

* `PUT /versions/{{app}}` or `POST /versions/{{app}}` with json like this:

```json
[
  {
    "name": "/v1",
    "weight": 2
  }
]
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
