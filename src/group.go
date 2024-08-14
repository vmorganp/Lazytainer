package main

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/cakturk/go-netstat/netstat"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"

	"github.com/google/gopacket"
	_ "github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

type LazyGroup struct {
	groupName           string   // the docker label associated with this group
	inactiveTimeout     uint16   // how many seconds of inactivity before group turns off
	minPacketThreshold  uint16   // minimum network traffic in packets before stop/pause occurs
	ignoreActiveClients bool     // determine container activity based on only packet count, ignoring connected client count.
	netInterface        string   // which network interface to watch traffic on. By default this is eth0 but can sometimes vary
	pollRate            uint16   // how frequently to poll traffic statistics
	ports               []uint16 // list of ports, which happens to also be a 16 bit range, how convenient!
	sleepMethod         string   // whether to stop or pause the container
}

var err error

func (lg LazyGroup) MainLoop() {
	// rxPacketCount is continuously updated by the getRxPackets goroutine
	var rxPacketCount int
	go lg.getRxPackets(&rxPacketCount)

	inactiveSeconds := 0
	// initialize a slice to keep track of recent network traffic
	rxHistory := make([]int, int(math.Ceil(float64(lg.inactiveTimeout/lg.pollRate))))
	sleepTime := time.Duration(lg.pollRate) * time.Second
	for {
		rxHistory = append(rxHistory[1:], rxPacketCount)
		if rxHistory[0] > rxHistory[len(rxHistory)-1] {
			rxHistory = make([]int, int(math.Ceil(float64(lg.inactiveTimeout/lg.pollRate))))
			debugLogger.Println("rx packets overflowed and reset")
		}
		// if the container is running, see if it needs to be stopped
		if lg.isGroupOn() {
			debugLogger.Println(rxHistory[len(rxHistory)-1]-rxHistory[0], "packets received in the last", lg.inactiveTimeout, "seconds")
			// if no clients are active on ports and threshold packets haven't been received in TIMEOUT secs
			if (lg.ignoreActiveClients || lg.getActiveClients() == 0) && rxHistory[0]+int(lg.minPacketThreshold) > rxHistory[len(rxHistory)-1] {
				// count up if no active clients
				inactiveSeconds = inactiveSeconds + int(lg.pollRate)
				fmt.Println(inactiveSeconds, "/", lg.inactiveTimeout, "seconds without an active client or sufficient traffic on running container")
				if inactiveSeconds >= int(lg.inactiveTimeout) {
					lg.stopContainers()
				}
			} else {
				inactiveSeconds = 0
			}
		} else {
			// if more than THRESHOLD rx in last RXHISTSECONDS seconds, start the container
			if rxHistory[0]+int(lg.minPacketThreshold) < rxHistory[len(rxHistory)-1] {
				inactiveSeconds = 0
				lg.startContainers()
			} else {
				debugLogger.Println(rxHistory[len(rxHistory)-1], "received out of", rxHistory[0]+int(lg.minPacketThreshold), "packets needed to restart container")
			}
		}
		time.Sleep(sleepTime)
		debugLogger.Println("/////////////////////////////////////////////////////////////////////////")
	}
}

func (lg LazyGroup) getContainers() []types.Container {
	filter := filters.NewArgs(filters.Arg("label", "lazytainer.group="+lg.groupName))
	containers, err := dockerClient.ContainerList(context.Background(), container.ListOptions{All: true, Filters: filter})
	check(err)

	return containers
}

func (lg LazyGroup) stopContainers() {
	debugLogger.Println("stopping container(s)")
	for _, c := range lg.getContainers() {
		if lg.sleepMethod == "stop" || lg.sleepMethod == "" {
			if err := dockerClient.ContainerStop(context.Background(), c.ID, container.StopOptions{}); err != nil {
				fmt.Printf("ERROR: Unable to stop container %s: %s\n", c.Names[0], err)
			} else {
				infoLogger.Println("stopped container ", c.Names[0])
			}
		} else if lg.sleepMethod == "pause" {
			if err := dockerClient.ContainerPause(context.Background(), c.ID); err != nil {
				fmt.Printf("ERROR: Unable to pause container %s: %s\n", c.Names[0], err)
			} else {
				infoLogger.Println("paused container ", c.Names[0])
			}
		}
	}
}

func (lg LazyGroup) startContainers() {
	debugLogger.Println("starting container(s)")
	for _, c := range lg.getContainers() {
		if lg.sleepMethod == "stop" || lg.sleepMethod == "" {
			if err := dockerClient.ContainerStart(context.Background(), c.ID, container.StartOptions{}); err != nil {
				fmt.Printf("ERROR: Unable to start container %s: %s\n", c.Names[0], err)
			} else {
				infoLogger.Println("started container ", c.Names[0])
			}
		} else if lg.sleepMethod == "pause" {
			if err := dockerClient.ContainerUnpause(context.Background(), c.ID); err != nil {
				fmt.Printf("ERROR: Unable to unpause container %s: %s\n", c.Names[0], err)
			} else {
				infoLogger.Println("unpaused container ", c.Names[0])
			}
		}
	}
}

func (lg LazyGroup) getRxPackets(packetCount *int) {
	handle, err := pcap.OpenLive(lg.netInterface, 1600, true,
		pcap.BlockForever) // TODO I have no idea if 1600 is a "good" number. It's what's in the example in the docs though
	check(err)
	defer handle.Close()

	// configure filter based on passed ports, looks like "port 80 or port 81 or etc..."
	var filter string
	for _, v := range lg.ports {
		filter += "port " + strconv.Itoa(int(v)) + " or "
	}
	filter = filter[0 : len(filter)-4]

	if err := handle.SetBPFFilter(filter); err != nil {
		panic(err)
	}

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for range packetSource.Packets() {
		// At some point this wraps around I think.
		// I have no idea when that point is or what the consequences of letting it happen are so I'm forcing it to be 1m
		*packetCount = (*packetCount + 1) % 1000000
		debugLogger.Println("group", lg.groupName, "received", *packetCount, "packets")
	}
}

func (lg LazyGroup) getActiveClients() int {
	// get active clients
	var allSocks []netstat.SockTabEntry

	// ipv4
	if ipv6Enabled {
		udpSocks, err := netstat.UDPSocks(netstat.NoopFilter)
		check(err)
		allSocks = append(allSocks, udpSocks...)
		tcpSocks, err := netstat.TCPSocks(netstat.NoopFilter)
		check(err)
		allSocks = append(allSocks, tcpSocks...)
	}

	// ipv6
	if ipv6Enabled {
		udp6Socks, err := netstat.UDP6Socks(netstat.NoopFilter)
		check(err)
		allSocks = append(allSocks, udp6Socks...)
		tcp6Socks, err := netstat.TCP6Socks(netstat.NoopFilter)
		check(err)
		allSocks = append(allSocks, tcp6Socks...)
	}

	activeClients := 0
	for _, socketEntry := range allSocks {
		if socketEntry.State.String() == "ESTABLISHED" {
			for _, aPort := range lg.ports {
				if strconv.FormatUint(uint64(aPort), 10) == fmt.Sprintf("%v", socketEntry.LocalAddr.Port) {
					activeClients++
				}
			}
		}
	}
	debugLogger.Println(activeClients, "active clients")
	return activeClients
}

func (lg LazyGroup) isGroupOn() bool {
	for _, c := range lg.getContainers() {
		if c.State == "running" {
			return true
		}
	}
	return false
}
