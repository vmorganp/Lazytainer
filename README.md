# Lazytainer - Lazy Load Containers
Putting your containers to sleep  
[![Docker](https://github.com/vmorganp/Lazytainer/actions/workflows/docker-publish.yml/badge.svg)](https://github.com/vmorganp/Lazytainer/actions/workflows/docker-publish.yml)
---

## How it works
Monitors network traffic for active connections and recieved packets  
If traffic looks to be idle, container stops  
If traffic is incoming to stopped container, container starts

## Want to test it?
```
$ git clone https://github.com/vmorganp/Lazytainer
$ cd Lazytainer
$ docker-compose up 
```

## Or put in your docker-compose.yaml
```
  lazytainer:
    container_name: lazytainer
    image: ghcr.io/vmorganp/lazytainer:master
    environment:
      - PORT=81                # comma separated list of ports...or just the one 
      - LABEL=lazytainer       # value of lazytainer.marker for other containers that lazytainer checks
      # - TIMEOUT=30           # OPTIONAL number of seconds to let container idle
      # - MINPACKETTHRESH=10   # OPTIONAL number of packets that must be recieved to keepalive/start container 
      # - POLLRATE=1           # OPTIONAL number of seconds to sleep between polls
      # - VERBOSE=true         # probably set this to false unless you're debugging or doing the initial demo
    ports:
      - 81:81
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro

  whoami1:
    container_name: whoami1
    image: containous/whoami
    command: --port 81 # make this run on the port passed through on lazytainer
    network_mode: service:lazytainer
    depends_on: 
      - lazytainer # wait for lazytainer to start before starting
    labels:
      - "lazytainer.marker=lazytainer" # required label to make it work
      - "lazytainer.sleepMethod=stop"       # can be either "stop" or "pause", or left blank for stop
```

## Configuration
### Notes
- Lazytainer does not "automatically" start and stop all of your containers. You must apply a label to them and proxy their traffic through the Lazytainer container.

### Environment Variables
| Variable        | Purpose                                                                                                    |
| --------------- | ---------------------------------------------------------------------------------------------------------- |
| PORT            | Port number(s) to listen for traffic on. If specifying multiple, they should be comma separated            |
| LABEL           | Value for label `lazytainer.marker` that lazytainer should use to determine which containers to start/stop |
| TIMEOUT         | Number of seconds container will be allowed to run with no traffic before it is stopped                    |
| MINPACKETTHRESH | Mimimum amount of recieved network packets to keep container alive                                         |
| POLLRATE        | Number of seconds to wait between polls of network transmission stats                                      |
| VERBOSE         | Whether or not to print noisier logs that may be useful for debugging                                      |
