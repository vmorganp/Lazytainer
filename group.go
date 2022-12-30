package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/cakturk/go-netstat/netstat"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

// TODO get aggregated logger and docker client to be global
type LazyGroup struct {
	groupName          string   // the docker label associated with this group
	inactiveTimeout    uint16   // how many seconds of inactivity before group turns off
	minPacketThreshold uint16   // Minimum network traffic in packets before stop/pause occurs
	netInterface       string   // which network interface to watch traffic on. By default this is eth0 but can sometimes vary
	pollRate           uint16   // how frequently to poll traffic statistics
	ports              []uint16 // List of ports, which happens to also be a 16 bit range, how convenient!
	stopMethod         string   // whether to stop or pause the container
}

func (lg LazyGroup) MainLoop() {
	// inactiveSeconds := 0
	sleepTime := time.Duration(lg.pollRate) * time.Second
	for {
		debugLogger.Println("tick")
		debugLogger.Println(lg.groupName, "group is waiting for", sleepTime.Seconds(), "second(s)")
		time.Sleep(sleepTime)
		lg.stopContainers()
		time.Sleep(time.Second * 5)
		lg.startContainers()
		infoLogger.Println(lg.getActiveClients())
	}
}

func (lg LazyGroup) getContainers() []types.Container {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv)
	check(err)
	filter := filters.NewArgs(filters.Arg("label", "lazytainer.groups="+lg.groupName))
	containers, err := dockerClient.ContainerList(context.Background(), types.ContainerListOptions{All: true, Filters: filter})
	check(err)

	return containers
}

func (lg LazyGroup) stopContainers() {
	fmt.Println("DEBUG: stopping container(s)")
	dockerClient, err := client.NewClientWithOpts(client.FromEnv)
	check(err)
	for _, c := range lg.getContainers() {
		if lg.stopMethod == "stop" || lg.stopMethod == "" {
			if err := dockerClient.ContainerStop(context.Background(), c.ID, nil); err != nil {
				fmt.Printf("ERROR: Unable to stop container %s: %s\n", c.Names[0], err)
			} else {
				fmt.Println("INFO: stopped container ", c.Names[0])
			}
		} else if lg.stopMethod == "pause" {
			if err := dockerClient.ContainerPause(context.Background(), c.ID); err != nil {
				fmt.Printf("ERROR: Unable to pause container %s: %s\n", c.Names[0], err)
			} else {
				fmt.Println("INFO: paused container ", c.Names[0])
			}
		}
	}
}

func (lg LazyGroup) startContainers() {
	fmt.Println("DEBUG: starting container(s)")
	dockerClient, err := client.NewClientWithOpts(client.FromEnv)
	check(err)
	for _, c := range lg.getContainers() {
		if lg.stopMethod == "stop" || lg.stopMethod == "" {
			if err := dockerClient.ContainerStart(context.Background(), c.ID, types.ContainerStartOptions{}); err != nil {
				fmt.Printf("ERROR: Unable to start container %s: %s\n", c.Names[0], err)
			} else {
				fmt.Println("INFO: started container ", c.Names[0])
			}
		} else if lg.stopMethod == "pause" {
			if err := dockerClient.ContainerUnpause(context.Background(), c.ID); err != nil {
				fmt.Printf("ERROR: Unable to unpause container %s: %s\n", c.Names[0], err)
			} else {
				fmt.Println("INFO: unpaused container ", c.Names[0])
			}
		}
	}
}

// TODO TODO TODO TODO TODO TODO TODO TODO TODO TODO TODO
// TODO TODO TODO TODO TODO TODO TODO TODO TODO TODO TODO
// 			THIS IS UBERHECKED FOR MANY TO ONE
// 			   IT STRAIGHT UP WILL NOT WORK
// TODO TODO TODO TODO TODO TODO TODO TODO TODO TODO TODO
// TODO TODO TODO TODO TODO TODO TODO cTODO TODO TODO TODO

// func (lg LazyGroup) getRxPackets() int {
// 	// get rx packets outside of the if bc we do it either way
// 	rx, err := os.ReadFile("/sys/class/net/" + lg.netInterface + "/statistics/rx_packets")
// 	check(err)
// 	rxPackets, err := strconv.Atoi(strings.TrimSpace(string(rx)))
// 	check(err)
// 	// if verbose { // TODO log levels
// 	// 	fmt.Println(rxPackets, "rx packets")
// 	// }
// 	return rxPackets
// }

func (lg LazyGroup) getActiveClients() int {
	// get active clients
	var allSocks []netstat.SockTabEntry
	udpSocks, err := netstat.UDPSocks(netstat.NoopFilter)
	check(err)
	udp6Socks, err := netstat.UDP6Socks(netstat.NoopFilter)
	check(err)
	tcpSocks, err := netstat.TCPSocks(netstat.NoopFilter)
	check(err)
	tcp6Socks, err := netstat.TCP6Socks(netstat.NoopFilter)
	check(err)

	activeClients := 0
	for _, socketEntry := range append(append(append(append(allSocks, udp6Socks...), tcp6Socks...), tcpSocks...), udpSocks...) {
		if socketEntry.State.String() == "ESTABLISHED" {
			for _, aPort := range lg.ports {
				if strconv.FormatUint(uint64(aPort), 10) == fmt.Sprintf("%v", socketEntry.LocalAddr.Port) {
					activeClients++
				}
			}
		}
	}
	// if verbose { // TODO log levels
	// 	fmt.Println(activeClients, "active clients")
	// }
	return activeClients
}
