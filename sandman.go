package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func main() {
	inactive_seconds := 0
	// label := os.Getenv("LABEL")
	inactive_timeout, err := strconv.Atoi(os.Getenv("TIMEOUT"))
	check(err)
	// tx
	// port := os.Getenv("PORT")
	rx_history := []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	for {
		// check if container is on
		container_state_on := is_container_on()

		// get rx packets outside of the if bc we do it either way
		rx, err := os.ReadFile("/sys/class/net/eth0/statistics/rx_packets")
		check(err)
		rx_packets, err := strconv.Atoi(strings.TrimSpace(string(rx)))
		check(err)
		rx_history = append(rx_history[1:], rx_packets)

		// if the container is running, see if it needs to be stopped
		if container_state_on {
			// get active clients
			out, err := exec.Command("/bin/sh", "-c", "netstat -n | grep ESTABLISHED | awk '{ print $4 }' | rev | cut -d: -f1| rev | grep -w \"$PORT\" | wc -l").Output() // todo make this handle multiple ports?
			check(err)
			active_clients, err := strconv.Atoi(strings.TrimSpace(string(out)))
			check(err)

			// log out the results
			println(active_clients, "active clients")
			println(rx_packets, "rx packets")
			fmt.Printf("%v rx history\n\n", rx_history)

			if active_clients == 0 && rx_history[0]+10 > rx_history[9] {
				// count up if we have no active clients
				// if no clients are active and less than 10 packets recieved in the last 10 seconds
				inactive_seconds++
				println(inactive_seconds, "seconds without an active client")
				if inactive_seconds > inactive_timeout {
					stop_containers()
				}
			}
		} else {
			// if more than 10 rx in last 10 seconds, start the container
			if rx_history[0]+10 < rx_history[9] {
				start_containers()
			}
		}

		time.Sleep(time.Second)
	}
}

type container struct {
	id    string
	state string
}

func newContainer(id string, state string) *container {
	c := container{id: id}
	c.state = state
	return &c
}

func get_containers() []container {
	containers := []container{}
	out, err := exec.Command("/bin/sh", "-c", "docker ps -a --no-trunc --filter label=\"com.sandman.marker=$LABEL\" --format \"{{.ID}} {{.State}}\"").Output() // todo make this handle multiple ports?
	check(err)
	fmt.Println(string(out))
	if strings.TrimSpace(string(out)) == "" {
		return nil
	}
	lines := strings.Split(string(strings.TrimSpace(string(out))), "\n")
	for _, s := range lines {
		containers = append(containers, container{strings.Split(s, " ")[0], strings.Split(s, " ")[1]})
	}
	return containers
}

func is_container_on() bool {
	containers := get_containers()
	for _, c := range containers {
		if c.state == "running" {
			return true
		}
	}
	return false
}

func stop_containers() {
	println("stopping container(s)")
	containers := get_containers()
	idString := ""
	for _, c := range containers {
		idString = idString + " " + c.id
	}

	out, err := exec.Command("/bin/sh", "-c", "docker stop "+idString).Output()
	check(err)
	fmt.Println(string(out))
}

func start_containers() {
	println("starting container(s)")
	containers := get_containers()
	idString := ""
	for _, c := range containers {
		idString = idString + " " + c.id
	}

	out, err := exec.Command("/bin/sh", "-c", "docker start "+idString).Output()
	check(err)
	fmt.Println(string(out))
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
