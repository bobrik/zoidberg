# Zoidberg

Zoidberg provides per app service discovery for Mesos with pluggable
discovery mechanisms. It allows you to shift traffic from version
to version in small percentages to ensure smooth deployments. It also
allows usual service discovery where it's up to framework how to
schedule tasks to avoid downtime.

## Compatible load balancers

* [zoidberg-nginx](https://github.com/bobrik/zoidberg-nginx):
supports upstream list updates without spawning new workers,
providing rich module ecosystem from nginx, including lua modules.

* [zoidberg-tcp](https://github.com/bobrik/zoidbergtcp):
zero-configuration TCP proxy that supports automatic dynamic service creation.

## Stability

Even though Zoidberg is deployed at scale (think 100s of Mesos slaves),
it is still stabilizing. Please see release notes before upgrading.

## Architecture

Zoidberg consists of several parts:

* Application finders responsible for finding apps.
* Load balancer finders responsible for finding load balancers.
* Load balancers are responsible for providing well known endpoints.
* Explorer is bounded with discoverer and responsible for version management.

It is possible to run several independent Zoidberg instances on a single cluster.
Each Zoidberg instance is only responsible for making sure that his group of
load balancers knows about current state of application. Different Zoidberg instances
can manage a single or completely independent groups of load balancers.

### Finders

Finder is a mechanism to discover load balancers and application tasks.

### Application finders

Application finders discover apps running on your cluster.

The following application finders are available:

* `marathon`
* `mesos`

You must specify application finder with `-application-finder` cli argument.

#### Mesos and Marathon finders

Both `marathon` and `mesos` finders are label based finders, which means
that they rely on labels to discover applications running on the cluster.

Make sure to use the following labels for your apps:

* `zoidberg_app_name` defines application name.
* `zoidberg_app_version` defines application version, defaults to `"1"`.
* `zoidberg_balanced_by` defines load balancer name for application.
* `zoidberg_meta_*` defines metadata labels for app, available in `meta`.

Arguments for `marathon` finder:

* `-application-finder-marathon-url` marathon url in `http://host:port[,host:port]` format.
* `-application-finder-marathon-balancer` balancer name to look for in `zoidberg_balanced_by` label.

Arguments for `mesos` finder:

* `-application-finder-mesos-masters` mesos masters in `http://host:port[,http://host:port]` format.
* `-application-finder-mesos-name` balancer name to look for in `zoidberg_balanced_by` label.

### Load balancer finders

Load balancer finders discover load balancers available on your cluster.

The following load balancer finders are available:

* `marathon`
* `mesos`
* `static`

You must specify load balancer finder with `-balancer-finder` cli argument.

#### Static finder

`static` finder has a predefined list of applications.

Arguments:

* `-balancer-finder-static-balancers` list of balancers in `host:port[,host:port]` format.

#### Mesos and Marathon finders

Both `marathon` and `mesos` finders are label based finders, which means
that they rely on labels to discover balancers available on the cluster.

Make sure to use the following labels for your apps that are load balancers:

* `zoidberg_balancer_for` defines load balancer name.

Arguments for `marathon` finder:

* `-balancer-finder-marathon-url` marathon url in `http://host:port[,host:port]` format.
* `-balancer-finder-marathon-name` balancer name to look for in `zoidberg_balancer_for` label.

Arguments for `mesos` finder:

* `-balancer-finder-mesos-masters` mesos masters in `http://host:port[,http://host:port]` format.
* `-balancer-finder-mesos-name` balancer name to look for in `zoidberg_balancer_for` label.

## Running

In addition to finder arguments, you also have to specify the following:

* `-name` Zoidberg instance name for identification.
* `-host` host to listen on for API.
* `-port` port to listen on for API.
* `-zk` Zookeeper connection string for state persistence.

Note that instead of cli arguments you can also use environment variables,
just drop the first `-`, replace and `-` with `_` and capitalize argument name.
For example, instead of specifying `-application-finder marathon` you could
set environment variable `APPLICATION_FINDER=marathon`.

Zoidberg is distributed as a docker image, below is an example how to run it
against Marathon running on a local Mesos cluster. Ttake a look at
[mesos-compose](https://github.com/bobrik/mesos-compose) to find out more about
running local mesos cluster.


```
docker run --rm -it --net host \
    -e HOST=0.0.0.0 \
    -e PORT=12345 \
    -e ZK=127.0.0.1:2181/zoidberg \
    -e APPLICATION_FINDER=marathon \
    -e APPLICATION_FINDER_MARATHON_URL=http://172.16.91.128:8080 \
    -e APPLICATION_FINDER_MARATHON_BALANCER=local \
    -e BALANCER_FINDER=static \
    -e BALANCER_FINDER_STATIC_BALANCERS=127.0.0.1:1234 \
    bobrik/zoidberg:0.5.0
```

### Zoidberg API

Zoidberg provides the next HTTP API:

* `PUT /versions/{{app}}` or `POST /versions/{{app}}` with json like this:

```json
{
  "1": {
    "weight": 2
  }
}
```

`{{app}}` in URL should be replaced with the name of an actual app.

* `GET /state` that returns full state (all set versions).

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
      "servers": [
        {
          "host": "192.168.0.7",
          "port": 31000,
          "ports": [31000],
          "version": "1"
        }
      ],
      "meta": {}
    }
  }
}
```

* `GET /_health` that returns 2xx code if everything looks good

## Why?

![zoidberg](zoidberg.jpg)
