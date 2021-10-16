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
    inactiveSeconds := 0
    // check if all of the necessary env vars are set, otherwise, use defaults
    // how long a container is allowed to have no traffic before being stopped
    inactiveTimeout, err := strconv.Atoi(os.Getenv("TIMEOUT"))
    if checkEnvFetchRecoverable(err) {
        fmt.Println("using default because env variable TIMEOUT not set ")
        inactiveTimeout = 30
    }

    // how many polls to keep around
    rxHistoryLength, err := strconv.Atoi(os.Getenv("RXHISTLENGTH"))
    if checkEnvFetchRecoverable(err) {
        fmt.Println("using default because env variable RXHISTLENGTH not set ")
        rxHistoryLength = 10
    }

    // number of packets required between first and last poll to keep container alive
    minPacketThreshold, err := strconv.Atoi(os.Getenv("MINPACKETTHRESH"))
    if checkEnvFetchRecoverable(err) {
        fmt.Println("using default because env variable MINPACKETTHRESH not set ")
        minPacketThreshold = 10
    }

    // ports to check for active connections
    ports := os.Getenv("PORT")
    check(err)
    if ports == "" {
        panic("you must set a port for this to work")
    }
    portsArray := strings.Split(string(strings.TrimSpace(string(ports))), ",")
    portsArg := ""
    for i, port := range portsArray {
        if i == 0 {
            portsArg += "'"
        }
        if i +1 < len(portsArray){
            portsArg += string(port)+"\\|"

        }else{
            portsArg += string(port)+"'"
        }
    }


    // how many seconds to wait in between polls
    pollRate, err := strconv.Atoi(os.Getenv("POLLRATE"))
    if checkEnvFetchRecoverable(err) {
        fmt.Println("using default because env variable POLLRATE not set ")
        pollRate = 1
    }

    rxHistory := make([]int, rxHistoryLength)

    for {
        // get rx packets outside of the if bc we do it either way
        rx, err := os.ReadFile("/sys/class/net/eth0/statistics/rx_packets")
        check(err)
        rxPackets, err := strconv.Atoi(strings.TrimSpace(string(rx)))
        check(err)
        rxHistory = append(rxHistory[1:], rxPackets)
        fmt.Println(rxPackets, "rx packets")

        // if the container is running, see if it needs to be stopped
        if isContainerOn() {
            // get active clients
            out, err := exec.Command("/bin/sh", "-c", "netstat -n | grep ESTABLISHED | awk '{ print $4 }' | rev | cut -d: -f1| rev | grep "+portsArg+" | wc -l").Output() // todo make this handle multiple ports?
            check(err)
            activeClients, err := strconv.Atoi(strings.TrimSpace(string(out)))
            check(err)

            // log out the results
            fmt.Println(activeClients, "active clients")

            if activeClients == 0 && rxHistory[0]+minPacketThreshold > rxHistory[len(rxHistory)-1] {
                // count up if we have no active clients
                // if no clients are active and less than 10 packets recieved in the last 10 seconds
                inactiveSeconds++
                fmt.Println(inactiveSeconds, "seconds without an active client")
                if inactiveSeconds >= inactiveTimeout {
                    stopContainers()
                }
            } else {
                inactiveSeconds = 0
            }
        } else {
            // if more than 10 rx in last 10 seconds, start the container
            if rxHistory[0]+minPacketThreshold < rxHistory[len(rxHistory)-1] {
                inactiveSeconds = 0
                startContainers()
            }
        }
        fmt.Println("//////////////////////////////////////////////////////////////////////////////////")
        for i := 0; i < pollRate; i++ {
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

func getContainers() []container {
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

func isContainerOn() bool {
    for _, c := range getContainers() {
        if c.state == "running" {
            return true
        }
    }
    return false
}

func stopContainers() {
    fmt.Println("stopping container(s)")
    idString := ""
    for _, c := range getContainers() {
        idString = idString + " " + c.id
    }

    out, err := exec.Command("/bin/sh", "-c", "docker stop "+idString).Output()
    check(err)
    fmt.Println(string(out))
}

func startContainers() {
    fmt.Println("starting container(s)")
    idString := ""
    for _, c := range getContainers() {
        idString = idString + " " + c.id
    }
    out, err := exec.Command("/bin/sh", "-c", "docker start "+idString).Output()
    check(err)
    fmt.Println(string(out))
}

func check(err error) {
    if err != nil {
        fmt.Println(err)
    }
}

func checkEnvFetchRecoverable(err error) bool {
    if err != nil {
        if strings.Contains(err.Error(), "strconv.Atoi: parsing \"\": invalid syntax") {
            return true
        }
        panic(err)
    }
    return false
}
