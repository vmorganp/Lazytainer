# Lazytainer
Putting your containers to sleep  

*I don't really wanna do the work today*

---

## How it works
### Lazy loading containers
monitor network traffic for active connections and recieved packets , 
if traffic looks to be idle, your container stops
if it looks like you're trying to access a stopped container, it starts

### Want to test it?
```
$ git clone https://github.com/vmorganp/Lazytainer
$ cd Lazytainer
$ docker-compose up 
```

## TODO
- support multiple ports
- test on common services
- docker security probably? this really shouldn't be getting exposed except to forward traffic so idk probably firewall all non listed ports?
- improve logging
- inevitable bugfixes
- ???
- profit
