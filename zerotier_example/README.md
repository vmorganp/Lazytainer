# Using Lazytainer over ZeroTier

We'll be using the zerotier docker image by [zyclonite](https://github.com/zyclonite/zerotier-docker) in a docker compose configuration.

Copy `docker-compose.yml` from this folder into an empty directory of your choosing.

Then run `docker compose up` after you have set the values to your liking.

## Connect to your ZeroTier network

Connect to your ZeroTier container:

```console
docker exec -it zerotier-one /bin/sh
```

Then connect your container to the network

```console
zerotier-cli join <yourZeroTierNetworkIDHere>
```

It should respond with `200 join OK` - Then authorize it in the [ZeroTier dashboard](https://my.zerotier.com/)

Whilst we're here, we should get the interface to listen on so that Lazytainer works over ZeroTier

```console
ifconfig
```

Copy the interface name that isn't `eth0` or `lo` - Mine was `ztukuxxqii`

Then run `exit` to return back to your host's terminal

## Secondary Configuration

Set `INTERFACE` in your configuration to the interface you got earlier

Then run `docker compose up` and Lazytainer should work as expected over ZeroTier.

If it works well for you, make sure to `Ctrl+C` to stop the containers and run `docker compose up -d` to run your configuration detached.
