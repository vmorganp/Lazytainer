# THIS NEEDS SOME DOC UPDATES BEFORE IT BECOMES A RELEASE
- Update the examples
  - verify that they work
  - use the new format
  - move them to the examples folder

- Update the main README
- Go through and comment new code
- Check code and make sure it isn't a disaster because I wrote it at 4 am

---

# Lazytainer - Lazy Load Containers
Putting your containers to sleep
[![Docker](https://github.com/vmorganp/Lazytainer/actions/workflows/docker-publish.yml/badge.svg)](https://github.com/vmorganp/Lazytainer/actions/workflows/docker-publish.yml)
---

## Quick Explanation
Monitors network traffic to containers. If there is traffic, the container runs, otherwise the container is stopped/paused. for more details check out [How it works](#how-it-works).

## Want to test it?
```
$ git clone https://github.com/vmorganp/Lazytainer
$ cd Lazytainer
$ docker-compose up
```
This will create 2 containers that you can reach through a third "lazytainer" container. 
To test this, go to your browser and navigate to `http://localhost:81`
If you close this tab and wait awhile you should see the container stop.
If you come back and refresh that page a few times after it has stopped, the container should restart.


## Configuration
### Note:
- Lazytainer does not "automatically" start and stop all of your containers. You must apply a label to them and proxy their traffic through the Lazytainer container.

### Groups 
Lazytainer starts and stops other containers in "groups" of one or more other containers. 
To assign a container to a lazytainer group, a label must be added. The label will look like 
```yaml
yourContainerThatWillSleep:
  ...
  labels:
    - "lazytainer.groups=yourGroupName"
```

Multiple containers may be in a group, such that they start and stop together.

To configure a group, there are some labels that must applied to the lazytainer container. 

```yaml
  lazytainer:
    ...
    ports: 
      - 81:81 # These ports are being passed through to the group members
      - 82:82
    # ...
    labels:
      # configuration for group 1
      # REQUIRED 
      - "lazytainer.group.group1.ports=81" # Network ports associated with this group
      # OPTIONAL
      - "lazytainer.group.group1.inactiveTimeout=30" # how long without sufficient network activity before sleeping
      - "lazytainer.group.group1.minPacketThreshold=30" # minimum amount of network packets for container to be on 
      - "lazytainer.group.group1.pollRate=30" # how often to check network activity
      - "lazytainer.group.group1.sleepMethod=pause" # can be either "stop" or "pause", or left blank for stop
      - "lazytainer.group.group1.netInterface=eth0" # network interface to listen on

      # configuration for group 2
      - "lazytainer.group.group2.ports=81,82" # You can use a comma separated list of ports as well, if you need more than one 
```

### Other configuration
#### Verbose Logging
If you would like more verbose logging, you can apply the environment variable `VERBOSE=true` to lazytainer like so
```yaml
  lazytainer:
    ...
    environment:
      - VERBOSE=true
```

#### volumes 
If using lazytainer, you MUST provide the following volume to lazytainer
```yaml
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
```

