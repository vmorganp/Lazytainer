package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"
)

func main() {
	label := os.Getenv("LABEL")
	// port := os.Getenv("PORT")
	inactive_seconds := 0
	host := true // true will represent on and false will represent off

	for {
		// if the host is turned on
		if host {
			out, err := exec.Command("/bin/sh", "-c", "netstat -n | grep ESTABLISHED | awk '{ print $4 }' | rev | cut -d: -f1| rev | grep -w \"$PORT\" | wc -l").Output()
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			fmt.Println("active clients: ", string(out))
			active_clients, _ := strconv.Atoi(string(out))
			if active_clients == 0 {
				inactive_seconds++
			}
			println(inactive_seconds, "seconds without an active client")
			if inactive_seconds > 10 {
				stop_containers(label)
				host = false
			}
		}

		if !host {
			println("time to check net stats for attempted connections")
			// do more checking here
			// check around in here to make sure the host doesn't magically restart and we have to deal with those again
			// really we should probably be checking that all the time anyways
			// out, err := exec.Command("/bin/sh", "-c", "netstat -n | grep ESTABLISHED | awk '{ print $4 }' | rev | cut -d: -f1| rev | grep -w \"$PORT\" | wc -l").Output()

		}

		time.Sleep(time.Second)
	}

	fmt.Println("hello world")
	// cmd := exec.Command(app, arg0, arg1, arg2, arg3)
	// stdout, err := cmd.Output()

	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	return
	// }

	// Print the output
	// fmt.Println(string(stdout))
}

func stop_containers(label string) {
	println("stopping contianer")

}

func start_containers(label string) {
	// docker ps -a --no-trunc --filter label="com.sandman.marker=sandman"
}
