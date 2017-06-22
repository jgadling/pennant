## Name TBD

Pennant is a powerful dark-launch utility designed for maximum flexibility and
performance.

### Getting Started

You'll need [govendor](https://github.com/kardianos/govendor) to manage dependencies. To install:

```
% go get -u github.com/kardianos/govendor
```

Using govendor:

```
# fetch dependencies and copy them into the vendor/ directory
% govendor get
# update packages from existing contents of your $GOPATH
% govendor update
```

Building and running pennant:

```
% brew install consul
% consul agent -dev -advertise 127.0.0.1
% git clone pennant.git
# note govendor wrapping the "build" command
% govendor build && ./pennant server
```

### API

| Method | Path | Description |
| --- | --- | --- |
| GET | /flags | List flags |
| GET | /flags/{name} | Get a flag's definition |
| DELETE | /flags/{name} | Delete a flag |
| POST | /flags | Create or update a flag |
| GET | /flagValue/{name} | Fetch en/disabled state of a flag, given a document |

### Example flag

```
{
  "name": "red_button",
  "description": "Makes the button on the home page red",
  "default": false,
  "policies": [
     {
      "comment": "Everybody whose username starts with 'foo'",
      "rules": "user_username =~ '^foo'"
     },
     {
      "comment": "and some volunteers",
      "rules": "user_id in (10, 11, 13)"
     },
     {
      "comment": "Also 10% of rando users",
      "rules": "pct(user_username) <= 10"
     }
  ]
}

```

### CLI


#### Create or update a flag

```
$ pennant update -f tests/data/flag1.json
- or -
$ pennant update '{"name":"red_button","description":....}'
- or -
$ cat tests/data/flag1.json | pennant update -

Name        Description                            DefaultValue
----------  -------------------------------------  ------------
red_button  Makes the button on the home page red  false

Rule                      Comment
------------------------  ------------------------------------------
user_username =~ '^foo'   Everybody whose username starts with 'foo'
user_id in (10, 11, 13)   and some volunteers
pct(user_username) <= 10  Also 10% of rando users
```

#### List flags

```
% pennant list
Name
----------
red_button
```

#### Get flag details

```
$ pennant show red_button
Name        Description                            DefaultValue
----------  -------------------------------------  ------------
red_button  Makes the button on the home page red  false

Rule                      Comment
------------------------  ------------------------------------------
user_username =~ '^foo'   Everybody whose username starts with 'foo'
user_id in (10, 11, 13)   and some volunteers
pct(user_username) <= 10  Also 10% of rando users
```

#### Check whether a flag is enabled

```
$ pennant value red_button '{"user_id": 10}'
- or -
$ pennant value red_button -f document.json
- or -
$ cat document.json | pennant value red_button -
Flag        Status
----------  -------
red_button  enabled
```

#### Delete a flag

```
$ pennant delete red_button
red_button deleted
```

#### Test a flag without a server

```
pennant test -f tests/data/flag1.json -d tests/data/data1.json
```


### Roadmap
V1 milestones:

 - ✓ Pluggable storage backends, ships with consul support
 - ✓ GRPC and REST query interfaces
 - ✓ REST flag management interfaces
 - ✓ Immediately updates flag cache when consul values change
 - ✓ Bundled percentage calculator
 - ✓ Supports arbitrary expressions for en/disabling flags
 - ✓ Client and server in single binary
 - Tests
 - Ships metrics to StatsD

V2:

 - More drivers - sqlite, etcd, filesystem?
 - Authentication
 - Prometheus compatible stats
 - Query results caching, more perf improvements
 - GRPC flag management interface

### Further reading on feature flags

- [http://featureflags.io/](http://featureflags.io/)
- [https://engineering.instagram.com/flexible-feature-control-at-instagram-a7d3417658df](https://engineering.instagram.com/flexible-feature-control-at-instagram-a7d3417658df)
