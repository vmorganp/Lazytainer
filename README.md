# Lazytainer - Lazy Load Containers
Putting your containers to sleep
[![Docker](https://github.com/vmorganp/Lazytainer/actions/workflows/docker-publish.yml/badge.svg)](https://github.com/vmorganp/Lazytainer/actions/workflows/docker-publish.yml)
---

## Quick Explanation
Monitors network traffic to containers. If there is traffic, the container runs, otherwise the container is stopped/paused. for more details check out the [Configuration](##Configuration) section.

## Want to test it?
1. Clone the project
    ```
    git clone https://github.com/vmorganp/Lazytainer
    cd Lazytainer
    ```
2. Start the stack
    ```sh
    docker-compose up
    ```
    This will create 2 containers that you can reach through a third "lazytainer" container
3. View the running container by navigating to its web ui at `http://localhost:81`. You should see some information about the container
4. Close the tab and wait until the logs say "stopped container"
6. Navigate again to `http://localhost:81`, it should be a dead page
7. Navigate to `http://localhost:81` several times, enough to generate some network traffic, and it should start
8. To clean up, run 
    ```sh
    docker-compose down
    ```

## Configuration
### Note:
Lazytainer does not "automatically" start and stop all of your containers. You must apply a label to them and proxy their traffic through the Lazytainer container.

### Examples
For examples of lazytainer in action, check out the [Examples](./examples/)

### Groups 
Lazytainer starts and stops other containers in "groups" of one or more other containers. 
To assign a container to a lazytainer group, a label must be added. The label will look like 
```yaml
yourContainerThatWillSleep:
  ...
  labels:
    - "lazytainer.group=yourGroupName"
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
      - "lazytainer.group.group1.inactiveTimeout=30"    # how long without sufficient network activity before sleeping
      - "lazytainer.group.group1.minPacketThreshold=30" # minimum amount of network packets for container to be on 
      - "lazytainer.group.group1.pollRate=30"           # how often to check network activity
      - "lazytainer.group.group1.sleepMethod=pause"     # can be either "stop" or "pause", or left blank for stop
      - "lazytainer.group.group1.netInterface=eth0"     # network interface to listen on

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

