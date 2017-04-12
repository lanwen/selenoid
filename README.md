# Selenoid
[![Build Status](https://travis-ci.org/aandryashin/selenoid.svg?branch=master)](https://travis-ci.org/aandryashin/selenoid)
[![Coverage](https://codecov.io/github/aandryashin/selenoid/coverage.svg)](https://codecov.io/gh/aandryashin/selenoid)
[![Go Report Card](https://goreportcard.com/badge/github.com/aandryashin/selenoid)](https://goreportcard.com/report/github.com/aandryashin/selenoid)
[![Release](https://img.shields.io/github/release/aandryashin/selenoid.svg)](https://github.com/aandryashin/selenoid/releases/latest)
[![GoDoc](https://godoc.org/github.com/aandryashin/selenoid?status.svg)](https://godoc.org/github.com/aandryashin/selenoid)

Selenoid is a powerful [Go](http://golang.org/) implementation of original [Selenium](http://github.com/SeleniumHQ/selenium) hub code. It is using Docker to launch browsers.

## Quick Start Guide

1) Install [Docker](https://docs.docker.com/engine/installation/)

2) Pull browser images, e.g.:
```
$ docker pull selenoid/firefox:latest
$ docker pull selenoid/chrome:latest
```

3) Pull Selenoid image:
```
$ docker pull aandryashin/selenoid:1.1.0
```

4) Create the following configuration file:
```bash
$ cat /etc/selenoid/browsers.json
{
    "firefox": {
      "default": "latest",
      "versions": {
        "latest": {
          "image": "selenoid/firefox:latest",
          "port": "4444",
          "path": "/wd/hub"
        }
      }
    },
    "chrome": {
      "default": "latest",
      "versions": {
        "latest": {
          "image": "selenoid/chrome:latest",
          "port": "4444"
        }
      }
    }
}
```

5) Run Selenoid container:
```
# docker run -d --name selenoid -p 4444:4444 -v /etc/selenoid:/etc/selenoid:ro -v /var/run/docker.sock:/var/run/docker.sock aandryashin/selenoid:1.1.0
```

6) Access Selenoid as regular Selenium hub:
```
http://localhost:4444/wd/hub
```

## Simple UI

Selenoid has standalone UI to show current quota, queue and browser usage: https://github.com/lanwen/selenoid-ui.  
