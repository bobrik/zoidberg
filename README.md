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
Two discoverers are provided with zoidberg:

* Marathon discoverer that takes information from marathon API
* Mesos discoverer that takes information from mesos task labels
and framework agnostic

In both provided discoverers both load balancers and application tasks
are scheduled on mesos.

### Load balancers

Load balancers are responsible for providing service endpoints
at well-known address and load balancing. Zoidberg requires balancers
to implement the next HTTP API:

* `PUT /upstreams` or `POST /upstreams` with json like this:

```json
[{"host":"somehost","port":123,"weight":5}]
```

In both provided discoverers balancers provide this API on endpoint
that is assigned by mesos schedulers.

### Explorer

Explorer is responsible for distributing state of mesos cluster to balancers.
It is also responsible for switching traffic between application versions.

Explorer provides the next HTTP API:

* `PUT /state` or `POST /state` with json like this:

```json
{"versions":{"/v1":{"name":"/v1","weight":2}}}
```

* `GET /state` that returns what was saved

* `GET /discovery` that returns json like this:

```json
{"balancers":[{"host":"192.168.59.103","port":31000}],"servers":{"/v1":[{"host":"192.168.59.103","port":31001}]}}
```

* `GET /_health` that returns 2xx code if everything looks good

## Why?

![zoidberg](zoidberg.jpg)
