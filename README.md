## Name TBD

Pennant is a powerful dark-launch utility designed for maximum flexibility and
performance.

### Getting Started

```
% brew install consul
% consul agent -dev -advertise 127.0.0.1
% git clone pennant.git
% go build && ./pennant server
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

TODO

To test whether a feature flag will be enabled given a flag definition and a
document:

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
 - Client and server in single binary
 - Ships metrics to StatsD

V2:

 - Authentication
 - Prometheus compatible stats
 - Query results caching, more perf improvements
 - GRPC flag management interface

### Further reading on feature flags

- [http://featureflags.io/](http://featureflags.io/)
- [https://engineering.instagram.com/flexible-feature-control-at-instagram-a7d3417658df](https://engineering.instagram.com/flexible-feature-control-at-instagram-a7d3417658df)
