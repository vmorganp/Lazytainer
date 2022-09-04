package main

import (
	"context"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cakturk/go-netstat/netstat"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

var (
	label      string
	portsArray []string

	ports              string
	inactiveTimeout    int
	minPacketThreshold int
	pollRate           int
	verbose            bool
)

func main() {
	setVarsFromEnv()
	inactiveSeconds := 0
	rxHistory := make([]int, int(math.Ceil(float64(inactiveTimeout/pollRate))))
	sleepTime := time.Duration(pollRate) * time.Second
	for {
		rxHistory = append(rxHistory[1:], getRxPackets())
		if rxHistory[0] > rxHistory[len(rxHistory)-1] {
			rxHistory = make([]int, int(math.Ceil(float64(inactiveTimeout/pollRate))))
			if verbose {
				fmt.Println("rx packets overflowed and reset")
			}
		}
		// if the container is running, see if it needs to be stopped
		if isContainerOn() {
			if verbose {
				fmt.Println(rxHistory[len(rxHistory)-1]-rxHistory[0], "packets received in the last", inactiveTimeout, "seconds")
			}
			// if no clients are active on ports and threshold packets haven't been received in TIMEOUT secs
			if getActiveClients() == 0 && rxHistory[0]+minPacketThreshold > rxHistory[len(rxHistory)-1] {
				// count up if no active clients
				inactiveSeconds = inactiveSeconds + pollRate
				fmt.Println(inactiveSeconds, "/", inactiveTimeout, "seconds without an active client or sufficient traffic on running container")
				if inactiveSeconds >= inactiveTimeout {
					stopContainers()
				}
			} else {
				inactiveSeconds = 0
			}
		} else {
			// if more than THRESHOLD rx in last RXHISTSECONDS seconds, start the container
			if rxHistory[0]+minPacketThreshold < rxHistory[len(rxHistory)-1] {
				inactiveSeconds = 0
				startContainers()
			} else {
				if verbose {
					fmt.Println(rxHistory[len(rxHistory)-1], "received out of", rxHistory[0]+minPacketThreshold, "packets needed to restart container")
				}
			}
		}
		time.Sleep(sleepTime)
		if verbose {
			fmt.Println("//////////////////////////////////////////////////////////////////////////////////")
		}
	}
}

func setVarsFromEnv() {
	label = os.Getenv("LABEL")
	if label == "" {
		panic("you must set env variable LABEL")
	}

	portsCSV := os.Getenv("PORT")
	if portsCSV == "" {
		panic("you must set env variable PORT")
	}
	// ports to check for active connections
	portsArray = strings.Split(string(strings.TrimSpace(string(portsCSV))), ",")
	ports = ""
	for i, port := range portsArray {
		if i == 0 {
			ports += "'"
		}
		if i+1 < len(portsArray) {
			ports += string(port) + "\\|"
		} else {
			ports += string(port) + "'"
		}
	}

	// logging level, should probably use a lib for this
	verboseString := os.Getenv("VERBOSE")
	if strings.ToLower(verboseString) == "true" {
		verbose = true
	}

	var err error

	// how long a container is allowed to have no traffic before being stopped
	inactiveTimeout, err = strconv.Atoi(os.Getenv("TIMEOUT"))
	if err != nil {
		if strings.Contains(err.Error(), "strconv.Atoi: parsing \"\": invalid syntax") {
			fmt.Println("using default 60 because env variable TIMEOUT not set ")
			inactiveTimeout = 60
		} else {
			panic(err)
		}
	}

	// number of packets required between first and last poll to keep container alive
	minPacketThreshold, err = strconv.Atoi(os.Getenv("MINPACKETTHRESH"))
	if err != nil {
		if strings.Contains(err.Error(), "strconv.Atoi: parsing \"\": invalid syntax") {
			fmt.Println("using default 10 because env variable MINPACKETTHRESH not set ")
			minPacketThreshold = 10
		} else {
			panic(err)
		}
	}

	// how many seconds to wait in between polls
	pollRate, err = strconv.Atoi(os.Getenv("POLLRATE"))
	if err != nil {
		if strings.Contains(err.Error(), "strconv.Atoi: parsing \"\": invalid syntax") {
			fmt.Println("using default 5 because env variable POLLRATE not set ")
			pollRate = 5
		} else {
			panic(err)
		}
	}
}

func getRxPackets() int {
	// get rx packets outside of the if bc we do it either way
	rx, err := os.ReadFile("/sys/class/net/eth0/statistics/rx_packets")
	check(err)
	rxPackets, err := strconv.Atoi(strings.TrimSpace(string(rx)))
	check(err)
	if verbose {
		fmt.Println(rxPackets, "rx packets")
	}
	return rxPackets
}

func getActiveClients() int {
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
			for _, aPort := range portsArray {
				if aPort == fmt.Sprintf("%v", socketEntry.LocalAddr.Port) {
					activeClients++
				}
			}
		}
	}
	if verbose {
		fmt.Println(activeClients, "active clients")
	}
	return activeClients
}

func getContainers() []types.Container {
	if label == "" {
		panic("environment variable LABEL must be set")
	}

	dockerClient, err := client.NewClientWithOpts(client.FromEnv)
	check(err)
	filter := filters.NewArgs(filters.Arg("label", "lazytainer.marker="+label))
	containers, err := dockerClient.ContainerList(context.Background(), types.ContainerListOptions{All: true, Filters: filter})
	check(err)

	// for _, container := range containers {
	// 	fmt.Printf("%s %s %s\n", container.ID[:10], container.Image, container.State)
	// }
	return containers
}

func isContainerOn() bool {
	for _, c := range getContainers() {
		if c.State == "running" {
			return true
		}
	}
	return false
}

func stopContainers() {
	fmt.Println("stopping container(s)")
	dockerClient, err := client.NewClientWithOpts(client.FromEnv)
	check(err)
	for _, c := range getContainers() {
		stopMethod := strings.ToLower(c.Labels["lazytainer.sleepMethod"])
		if stopMethod == "stop" || stopMethod == "" {
			if err := dockerClient.ContainerStop(context.Background(), c.ID, nil); err != nil {
				fmt.Printf("Unable to stop container %s: %s\n", c.Names[0], err)
			} else {
				fmt.Println("stopped container ", c.Names[0])
			}
		} else if stopMethod == "pause" {
			if err := dockerClient.ContainerPause(context.Background(), c.ID); err != nil {
				fmt.Printf("Unable to pause container %s: %s\n", c.Names[0], err)
			} else {
				fmt.Println("paused container ", c.Names[0])
			}
		}
	}
}

func startContainers() {
	fmt.Println("starting container(s)")
	dockerClient, err := client.NewClientWithOpts(client.FromEnv)
	check(err)
	for _, c := range getContainers() {
		stopMethod := strings.ToLower(c.Labels["lazytainer.sleepMethod"])
		if stopMethod == "stop" || stopMethod == "" {
			if err := dockerClient.ContainerStart(context.Background(), c.ID, types.ContainerStartOptions{}); err != nil {
				fmt.Printf("Unable to start container %s: %s\n", c.Names[0], err)
			} else {
				fmt.Println("started container ", c.Names[0])
			}
		} else if stopMethod == "pause" {
			if err := dockerClient.ContainerUnpause(context.Background(), c.ID); err != nil {
				fmt.Printf("Unable to unpause container %s: %s\n", c.Names[0], err)
			} else {
				fmt.Println("unpaused container ", c.Names[0])
			}
		}
	}
}

func check(err error) {
	if err != nil {
		fmt.Println(err)
	}
}
