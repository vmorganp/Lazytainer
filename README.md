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
    DOCKER_BUILDKIT=1 docker-compose up --build
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
Lazytainer starts and stops other containers in "groups" of one or more other containers. To assign a container to a lazytainer group, a label must be added. The label will look like this.

```yaml
yourContainerThatWillSleep:
  # ... configuration omitted for brevity
  labels:
    - "lazytainer.group=<yourGroupName>"
```

To configure a group, add labels to the lazytainer container like this. Note that each is required to have a port(s) specified. These ports must also be forwarded on the lazytainer container
```yaml
  lazytainer:
    # ... configuration omitted for brevity
    ports: 
      - 81:81 # used by group1 and group2
      - 82:82 # used by group2
    labels:
      # Configuration items are formatted like this
      - "lazytainer.group.<yourGroupName>.<property>=value"
      # configuration for group 1
      - "lazytainer.group.group1.ports=81"
      # configuration for group 2
      - "lazytainer.group.group2.ports=81,82"
```

Group properties that can be changed include:

| Name               | description                                                                            | required | default |
| ------------------ | -------------------------------------------------------------------------------------- | -------- | ------- |
| ports              | Network ports associated with a group, can be comma separated                          | Yes      | n/a     |
| inactiveTimeout    | Time (seconds) before container is stopped when there is insufficient network activity | No       | 30      |
| minPacketThreshold | Minimum count of network packets for container to be on                                | No       | 30      |
| pollRate           | How frequently (seconds) to check network activity                                     | No       | 30      |
| sleepMethod        | How to put the container to sleep. Can be `stop` or `pause`                            | No       | `stop`  |
| netInterface       | Network interface to listen on                                                         | No       | `eth0`  |

### Additional Configuration
#### Verbose Logging
If you would like more verbose logging, you can apply the environment variable `VERBOSE=true` to lazytainer like so
```yaml
  lazytainer:
    # ... configuration omitted for brevity
    environment:
      - VERBOSE=true
```

#### Volumes 
If using lazytainer, you MUST provide the following volume to lazytainer
```yaml
  lazytainer:
    # ... configuration omitted for brevity
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
```
