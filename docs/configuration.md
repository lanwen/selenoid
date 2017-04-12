## Configuration

### Flags

The following flags are supported by ```selenoid``` command:
```
-conf string
    Browsers configuration file (default "config/browsers.json")
-disable-docker
    Disable docker support
-limit int
    Simultaneous container runs (default 5)
-listen string
    Network address to accept connections (default ":4444")
-log-conf string
    Container logging configuration file (default "config/container-logs.json")
-timeout duration
    Session idle timeout in time.Duration format (default 1m0s)
```
For example:
```
$ ./selenoid -conf /my/custom/browsers.json -limit 10
```
When using Selenoid inside Docker container these flags are passed like the following:
```
# docker run -d --name selenoid -p 4444:4444 -v /etc/selenoid:/etc/selenoid:ro -v /var/run/docker.sock:/var/run/docker.sock aandryashin/selenoid:1.1.0 -conf /my/custom/browsers.json -limit 10
```

### Simultaneously Running Containers
Total number of simultaneously running containers (adjusted via ```-limit``` flag) depends on your host machine hardware. Our experience shows that depending on your tests the recommended limit is something like: ```1.5-2.0 x numCores```, where ```numCores``` is total number of cores on your host machine.

### Browsers Configuration File

Selenoid uses simple JSON configuration files of the following format (we use **#** for comments here):
```
{
    "firefox": { # Browser name
      "default": "46.0", # Default browser version
      "versions": { # A list of available browser versions
        "46.0": { # Version name
          "image": "selenoid/firefox:46.0", # Image name or driver binary command
          "port": "4444", # Port to proxy connections to, see below
          "tmpfs": {"/tmp": "size=512m"}, # Optional. Add in memory filesystem (tmpfs) to container, see below
          "path" : "/wd/hub" # Optional. Path relative to / where we request a new session, see below
        },
        "50.0" :{
            # ...
        }
      }
    },
    "chrome": {
        # ...
    }
}
```
This file represents a mapping between a list of supported browser versions and Docker container images or driver binaries.
#### Browser Name and Version
Browser name and version are just strings that are matched against Selenium desired capabilities: browserName and version. If no version capability is present default version is used. When there is no exact version match we also try match by prefix. That means version string in JSON should start with version string from capabilities. The following request matches...
```
versionFromConfig = 46.0
versionFromCapabilities = 46 # 46.0 starts with 46
```
... and the following does not:
```
versionFromConfig = 46.0
versionFromCapabilities = 46.1 # 46.0 does not start with 46.1
```
#### Image
Image by default is a string with container specification in Docker format (hub.example.com/project/image:tag). The following image names are valid:
```
my-internal-docker-hub.example.com/selenoid/firefox:46.0 # This comes from internal Docker hub
selenoid/firefox:46.0 # This is downloaded from hub.docker.com
```
If you wish to use a standalone binary instead of Docker container, then image field should contain command specification in square brackets:
```
"46.0": { # Version name
    "image": ["/usr/bin/mybinary", "-arg1", "foo", "-arg2", "bar", "-arg3"],
    "port": ...
    "tmpfs": ...
    "path" : ...
},
```
Selenoid proxies connections to either Selenium server or standalone driver binary. Depending on operating system both can be packaged inside Docker container.

#### Port, Tmpfs and Path
You should use **port** field to specify the real port that Selenium server or driver will listen on. For Docker containers this is a port inside container. **tmpfs** and **path** fields are optional. You may probably know that moving browser cache to in-memory filesystem ([tmpfs](https://en.wikipedia.org/wiki/Tmpfs)) can dramatically improve its performance. Selenoid can automatically attach one or more in-memory filesystems as volumes to Docker container being run. To achieve this define one or more mount points and their respective sizes in optional **tmpfs** field:
```
"46.0": { # Version name
    "image": ...
    "port": ...
    "tmpfs": {"/tmp": "size=512m", "/var": "size=128m"},
    "path" : ...
},
```
The last field - **path** is needed to specify relative path to the URL where a new session is created (default is **/**).

### Timezone

When used in Docker container Selenoid will have timezone set to UTC. To set custom timezone pass TZ environment variable to Docker:
```
$ docker run -d --name selenoid -p 4444:4444 -e TZ=Europe/Moscow -v /etc/selenoid:/etc/selenoid:ro -v /var/run/docker.sock:/var/run/docker.sock aandryashin/selenoid:1.1.0
```

### Logging Configuration File
By default Docker container logs are saved to host machine hard drive. When using Selenoid for local development that's ok. But in big Selenium cluster you may want to send logs to some centralized storage like [Logstash](https://www.elastic.co/products/logstash) or [Graylog](https://www.graylog.org/). Docker provides such functionality by so-called [logging drivers](https://docs.docker.com/engine/admin/logging/overview/). Selenoid logging configuration file allows to specify which logging driver to use globally for all started Docker containers with browsers. Configuration file has the following format:
```
{
    "Type" : "<driver-type>",
    "Config" : {
      "key1" : "value1",
      "key2" : "value2"
    }
}
```
Here ```<driver-type>``` - is a supported Docker logging driver type like ```syslog```, ```journald``` or ```awslogs```. ```Config``` is a list of key-value pairs used to configure selected driver. For example these Docker logging parameters...
```
--log-driver=syslog --log-opt syslog-address=tcp://192.168.0.42:123 --log-opt syslog-facility=daemon
```
... are equivalent to the following Selenoid logging configuration:
```
{
    "Type" : "syslog",
    "Config" : {
      "syslog-address" : "tcp://192.168.0.42:123",
      "syslog-facility" : "daemon"
    }
}
```

### Reloading Configuration

To reload configuration without restart send SIGHUP:
```
# kill -HUP <pid> # Use only one of these commands!!
# docker kill -s HUP <container-id-or-name>
```
