# Lazy Load Docker Minecraft Server

## Startup

```
git clone https://github.com/vmorganp/Lazytainer
cd Lazytainer/examples/minecraft
docker-compose up
```

## Watch the magic happen

After a configurable period of no activity the server should stop
if you generate some traffic by trying to connect to the instance or running a command like

`telnet localhost 25565`

a few times

you should see the minecraft container automatically restart itself
