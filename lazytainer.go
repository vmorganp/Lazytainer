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
    // check if all of the necessary env vars are set, otherwise, use defaults
    // how long a container is allowed to have no traffic before being stopped
    inactive_timeout, err := strconv.Atoi(os.Getenv("TIMEOUT"))
    if check_env_fetch_recoverable(err) {
        fmt.Println("using default because env variable TIMEOUT not set ")
        inactive_timeout = 30
    }

    // how many polls to keep around
    rx_history_length, err := strconv.Atoi(os.Getenv("RXHISTLENGTH"))
    if check_env_fetch_recoverable(err) {
        fmt.Println("using default because env variable RXHISTLENGTH not set ")
        rx_history_length = 10
    }

    // number of packets required between first and last poll to keep container alive
    min_packet_threshhold, err := strconv.Atoi(os.Getenv("MINPACKETTHRESH"))
    if check_env_fetch_recoverable(err) {
        fmt.Println("using default because env variable MINPACKETTHRESH not set ")
        min_packet_threshhold = 10
    }

    // ports to check for active connections
    ports := os.Getenv("PORT")
    check(err)
    if ports == "" {
        panic("you must set a port for this to work")
    }
    ports_arr := strings.Split(string(strings.TrimSpace(string(ports))), ",")
    portsarg := ""
    for i, port := range ports_arr {
        if i == 0 {
            portsarg += "'"
        }
        if i +1 < len(ports_arr){
            portsarg += string(port)+"\\|"

        }else{
            portsarg += string(port)+"'"
        }
    }


    // how many seconds to wait in between polls
    poll_rate, err := strconv.Atoi(os.Getenv("POLLRATE"))
    if check_env_fetch_recoverable(err) {
        fmt.Println("using default because env variable POLLRATE not set ")
        poll_rate = 1
    }

    rx_history := make([]int, rx_history_length)

    for {
        // check if container is on
        container_state_on := is_container_on()

        // get rx packets outside of the if bc we do it either way
        rx, err := os.ReadFile("/sys/class/net/eth0/statistics/rx_packets")
        check(err)
        rx_packets, err := strconv.Atoi(strings.TrimSpace(string(rx)))
        check(err)
        rx_history = append(rx_history[1:], rx_packets)
        fmt.Println(rx_packets, "rx packets")

        // if the container is running, see if it needs to be stopped
        if container_state_on {
            // get active clients
            out, err := exec.Command("/bin/sh", "-c", "netstat -n | grep ESTABLISHED | awk '{ print $4 }' | rev | cut -d: -f1| rev | grep "+portsarg+" | wc -l").Output() // todo make this handle multiple ports?
            check(err)
            active_clients, err := strconv.Atoi(strings.TrimSpace(string(out)))
            check(err)

            // log out the results
            fmt.Println(active_clients, "active clients")
            // println(rx_packets, "rx packets")
            // fmt.Printf("%v rx history", rx_history)

            if active_clients == 0 && rx_history[0]+min_packet_threshhold > rx_history[len(rx_history)-1] {
                // count up if we have no active clients
                // if no clients are active and less than 10 packets recieved in the last 10 seconds
                inactive_seconds++
                fmt.Println(inactive_seconds, "seconds without an active client")
                if inactive_seconds > inactive_timeout {
                    stop_containers()
                }
            } else {
                inactive_seconds = 0
            }
        } else {
            // if more than 10 rx in last 10 seconds, start the container
            if rx_history[0]+min_packet_threshhold < rx_history[len(rx_history)-1] {
                inactive_seconds = 0
                start_containers()
            }
        }
        fmt.Println("//////////////////////////////////////////////////////////////////////////////////")
        for i := 0; i < poll_rate; i++ {
            time.Sleep(time.Second)
        }
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
    out, err := exec.Command("/bin/sh", "-c", "docker ps -a --no-trunc --filter label=\"lazytainer.marker=$LABEL\" --format \"{{.ID}} {{.State}}\"").Output() // todo make this handle multiple ports?
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
    fmt.Println("stopping container(s)")
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
    fmt.Println("starting container(s)")
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
        fmt.Println(err)
        // panic(err)
    }
}

func check_env_fetch_recoverable(err error) bool {
    if err != nil {
        if strings.Contains(err.Error(), "strconv.Atoi: parsing \"\": invalid syntax") {
            return true
        }
        panic(err)
    }
    return false
}
