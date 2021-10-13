# Sandman
Putting your containers to sleep  

*man me a sand*

---

## How it works
### Lazy loading containers
monitor network traffic for active connections and recieved packets , 
if traffic looks to be idle, your container stops
if it looks like you're trying to access a stopped container, it starts

### Want to test it?
```
$ git clone https://github.com/vmorganp/sandman
$ cd sandman
$ docker-compose up -d --build
```

## TODO
- support multiple ports
- test on common services
- better time adjustment
- docker security probably? this really shouldn't be getting exposed except to forward traffic so idk probably firewall all non listed ports?
- improve logging
- inevitable bugfixes
- ???
- profit
