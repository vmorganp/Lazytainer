# Lazytainer
Putting your containers to sleep  

*I don't really wanna do the work today*

---

## How it works
monitors network traffic for active connections and recieved packets  
if traffic looks to be idle, container stops
if traffic is incoming to stopped container, container starts

## Want to test it?
```
$ git clone https://github.com/vmorganp/Lazytainer
$ cd Lazytainer
$ docker-compose up 
```

## Or put in your docker compose
```
  lazytainer:
    container_name: lazytainer
    image: ghcr.io/vmorganp/lazytainer:master
    environment:
      - PORT=81           # comma separated list of ports...or just the one 
      - LABEL=lazytainer     # value of lazytainer.marker for other containers that lazytainer checks
      # - TIMEOUT=30           # OPTIONAL number of seconds to let container idle
      # - MINPACKETTHRESH=10   # OPTIONAL number of packets that must be recieved to keepalive/start container 
      # - POLLRATE=1           # OPTIONAL number of packets that must be recieved to keepalive/start container 
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
```