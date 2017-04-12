## Usage statistics

Selenoid calculates usage statistics that can be accessed with HTTP request:
```bash
$ curl http://localhost:4444/status
{
    "total": 80,
    "used": 14,
    "queued": 0,
    "pending": 1,
    "browsers": {
      "firefox": {
        "46.0": {
          "user1": 5,
          "user2": 6
        },
        "48.0": {
          "user2": 3
        }
      }
    }
}
```
Users are extracted from basic HTTP authentication headers.

### Sending statistics to external systems

To send Selenoid statistics described in previous section you can use [Telegraf](https://github.com/influxdata/telegraf). For example to send status to [Graphite](https://github.com/graphite-project):

1) Pull latest Telegraf Docker container:
```
# docker pull telegraf:alpine
```
2) Generate configuration file:
```
# mkdir -p /etc/telegraf
# docker run --rm telegraf:alpine --input-filter httpjson --output-filter graphite config > /etc/telegraf/telegraf.conf
```
3) Edit file like the following (three dots mean some not shown lines):
```go
...
[agent]
interval = "10s" # <- adjust this if needed
...
[[outputs.graphite]]
...
servers = ["my-graphite-host.example.com:2024"] # <- adjust host and port
...
prefix = "one_min" # <- something before hostname, can be blank
...
template = "host.measurement.field"
...
[[inputs.httpjson]]
...
name = "selenoid"
...
servers = [
"http://localhost:4444/status" # <- this is localhost only if Telegraf is run on Selenoid host
]
```
4) Start Telegraf container:
```
# docker run --net host -d --name telegraf -v /etc/telegraf:/etc telegraf:alpine --config /etc/telegraf.conf
```
Metrics will be sent to `my-graphite-host.example.com:2024`.

Metric names will be like the following:
```
one_min.selenoid_example_com.httpjson_selenoid.pending
one_min.selenoid_example_com.httpjson_selenoid.queued
one_min.selenoid_example_com.httpjson_selenoid.total
...
```
