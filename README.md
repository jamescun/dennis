# DENNIS

_Pronounced DI-N-ES, DENNIS EXTERNAL NETWORK NAME INQUIRY SYSTEM._

DENNIS is a web-based utility to query multiple DNS resolvers for the same name.

**Rationale:** I maintain multiple DNS servers whose configuration _should_ be identical, for both resolution and authoritative for internal domains in a split-horizon setup. This utility lets me view the resolution of all these DNS servers in one place, to identify configuration mismatches, as well as watch cached results be evicted.


## Table of Contents

- [Installation](#installation)
- [Running](#running)
- [Configuration](#configuration)
  - [Logging](#logging)
  - [Listen](#listen)
  - [Resolvers](#resolvers)
  - [Database](#database)
    - [File](#file)
	- [PostgreSQL](#postgresql)
	- [Redis](#redis)


## Installation

Either download a [pre-compiled binary or DEB/RPM package](https://github.com/jamescun/dennis/releases), or build from source using:

```sh
go install github.com/jamescun/dennis@latest
```

Alternatively, a container is built from [Dockerfile](Dockerfile) and published as [james/dennis](https://hub.docker.com/r/james/dennis).


## Running

If you compiled yourself or downloaded a binary, you may simply run:

```sh
./dennis --config path_to_config.yml
```

If you installed a DEB/RPM package, these include a systemd service and an example configuration file at `/etc/dennis/config.yml`, you can start with:

```sh
systemctl start dennis
```

If you are using Docker, you may run DENNIS like:

```sh
docker run --name dennis -p 8080:8080 -v ./config.yml:/etc/dennis/config.yml -v ./data:/data james/dennis:1.0.0
```

This will mount your local `config.yml` into the container as `/etc/dennis/config.yml` (the default path), mount the local directory `data/` as `/data`, and expose the DENNIS server at port 8080 on your machine.

You can also use the [docker-compose.yml](docker-compose.yml) file.


## Configuration

DENNIS is configured using a JSON or YAML configuration file. An example configuration file can be seen in [config.example.yml](config.example.yml).

This full configuration specification can be found in code at [app/config/config.go](app/config/config.go).


### Logging

The `logging` section configures how DENNIS logs.

| name  | type | required | description                                     |
| ----- | ---- | -------- | ----------------------------------------------- |
| debug | bool | false    | enable debug logging, default false             |
| json  | bool | false    | log using machine readable JSON instead of text |

**Example:**

```yaml
logging:
  debug: false
  json: true
```


### Listen

The `listen` section configures how the integrated web server in DENNIS will accept connections.

| name | type   | required | description                                 |
| ---- | ------ | -------- | ------------------------------------------- |
| addr | string | true     | `host:port` for the web server to listen on |

**Example:**

```yaml
listen:
  addr: localhost:8080
```


### Resolvers

The `resolvers` section configures one-or-more upstream DNS resolver servers that DENNIS will query for their view on a requested record.

It is an array of resolver configurations, and at least one resolver is required.

| name | type   | required | description                             |
| ---- | ------ | -------- | --------------------------------------- |
| name | string | true     | name of resolver as displayed in the UI |
| addr | string | true     | `host:port` for the DNS resolver        |

**Example:**

```yaml
resolvers:
- name: CloudFlare
  addr: 1.1.1.1:53
- name: Google DNS
  addr: 8.8.4.4:53
```


### Database

The `db` section configures where DENNIS stores requested queries and the results to those queries.

At most one of the below sections must be configured.

#### File

The `file` database backend uses a JSON file in the local filesystem to store queries and their results.

Care is taken to lock around read/write cycles to this file, but it is still only suitable for evaluation and small deployments.

| name | type   | required | description                               |
| ---- | ------ | -------- | ----------------------------------------- |
| path | string | true     | path to file to store queries and results |

**Example:**

```yaml
db:
  file:
    path: /tmp/dennis.json
```

#### PostgreSQL

The `postgres` database backend uses a PostgreSQL database to store queries and their results.

Once given a database to connect to, DENNIS will apply it's migrations to create the necessary tables.

| name | type   | required | description                                              |
| ---- | ------ | -------- | -------------------------------------------------------- |
| url  | string | true     | libpq-compatible connection string for PostgreSQL server |

See the [libpq](https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING-URIS) documentation for a description of the values supported for a connection url. Also see the [pgxpool](https://pkg.go.dev/github.com/jackc/pgx/v5@v5.8.0/pgxpool#ParseConfig) documentation, the PostgreSQL driver user, for it's handing of connection url.

**Example:**

```yaml
db:
  postgres:
    url: postgres://username:password@localhost/dennis?sslmode=disable
```


#### Redis

The `redis` database backend uses a Redis in-memory database to store queries and their results.

Internally this database backend uses the JSON key type.

| name     | type   | required | description                                   |
| -------- | ------ | -------- | --------------------------------------------- |
| addr     | string | true     | `host:port` of the Redis server to connect to |
| db       | int    | false    | id of database to use, default `0`            |
| username | string | false    | if authentication is enabled, username to use |
| password | string | false    | if authentication is enabled, password to use |

**Example:**

```yaml
db:
  redis:
    addr: localhost:6379
	username: dennis
	password: changeme
```
