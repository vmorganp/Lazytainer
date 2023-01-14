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

## Or put in your docker-compose.yaml
```
  lazytainer:
    container_name: lazytainer
    image: ghcr.io/vmorganp/lazytainer:master
    environment:
      - PORT=81               # comma separated list of ports...or just the one
      - LABEL=lazytainer      # value of lazytainer.marker for other containers that lazytainer checks
      # - TIMEOUT=30          # OPTIONAL number of seconds to let container idle
      # - MINPACKETTHRESH=10  # OPTIONAL number of packets that must be recieved to keepalive/start container
      # - POLLRATE=1          # OPTIONAL number of seconds to sleep between polls
      # - VERBOSE=true        # probably set this to false unless you're debugging or doing the initial demo
      # - INTERFACE=eth0      # OPTIONAL interface to listen on - use eth0 in most cases
    ports:
      - 81:81
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro

  whoami1:
    container_name: whoami1
    image: containous/whoami
    # configuring service ports is container specific. Look up how to do this on your service of choice
    command: --port 81  # make this run on the port passed through on lazytainer
    network_mode: service:lazytainer
    depends_on:
      - lazytainer # wait for lazytainer to start before starting
    labels:
      - "lazytainer.marker=lazytainer"  # required label to make it work
      - "lazytainer.sleepMethod=stop"   # can be either "stop" or "pause", or left blank for stop
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
| MINPACKETTHRESH | Minimum amount of received network packets to keep container alive                                         |
| POLLRATE        | Number of seconds to wait between polls of network transmission stats                                      |
| VERBOSE         | Whether or not to print noisier logs that may be useful for debugging                                      |
| INTERFACE       | What interface to check for received packets on                                                            |

## How it works
Lazytainer sits between users and the containers they are accessing, and acts as a network proxy.

Lazytainer checks to see if $MINPACKETTHRESH number of packets have been received in $TIMEOUT number of seconds. If the number of packets is above $MINPACKETTHRESH the container(s) with the label will ALL start/remain on depending on prior state. If the number of packets is less than $MINPACKETTHRESH, the container(s) with the label wil ALL stop/pause or remain stopped/pause depending on prior state and configuration.

If you use a reverse proxy like Caddy, NGINX, Traefik, or others, you can still point your reverse proxy of choice to your service. Instead of pointing directly at your service, you must instead point your reverse proxy to lazytainer, which will then pass your traffic to your service container.
