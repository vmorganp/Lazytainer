# Lazy Load Docker Satisfactory Server

## Startup

```
git clone https://github.com/vmorganp/Lazytainer
cd Lazytainer/examples/satisfactory
docker compose up
```

or

#### Deploy with Portainer, etc

Copy contents of docker-compose.yml into a stack, it should automatically deploy

## Notes

- "lazytainer.group.satisfactory.inactiveTimeout=120"

This may need to be adjusted based on your physical hardware. If you have slower hardware, the server client may not start with enough time to accept clients and create additional traffic.
In my experience, players can expect a 45 second delay after opening the Satisfactory client and navigating to the server manager before the server actually accepts clients. From the time clients are accepted, this gives players about a minute and a half to login before the container will try to shutdown again.
This could very well change based on hardware specifications, you may need to adjust.

Don't forget to portforward
