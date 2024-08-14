package main

import (
	"context"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

// log := logger()
var infoLogger *log.Logger
var debugLogger *log.Logger
var dockerClient *client.Client
var ipv6Enabled bool
var ipv4Enabled bool

func main() {
	flags := log.LstdFlags | log.Lshortfile
	infoLogger = log.New(os.Stdout, "INFO: ", flags)
	debugLogger = log.New(os.Stdout, "DEBUG: ", flags)

	// if the verbose flag isn't set to true, don't log debug logs
	verbose, verboseFlagSet := os.LookupEnv("VERBOSE")
	if !verboseFlagSet || strings.ToLower(verbose) != "true" {
		debugLogger.SetOutput(io.Discard)
	}

	// if the IPV6_DISABLED flag isn't set to true, leave ipv6 enabled
	envIpv6, existsEnvDisableIpv6 := os.LookupEnv("IPV6_DISABLED")
	if existsEnvDisableIpv6 && strings.ToLower(envIpv6) != "true" {
		ipv6Enabled = true
	} else {
		ipv6Enabled = false
	}

	// if the IPV4_DISABLED flag isn't set to true, leave ipv4 enabled
	envIpv4, existsEnvDisableIpv4 := os.LookupEnv("IPV4_DISABLED")
	if existsEnvDisableIpv4 && strings.ToLower(envIpv4) != "true" {
		ipv4Enabled = true
	} else {
		ipv4Enabled = false
	}

	// configure groups. eventually it might be nice to have file based config as well.
	groups := configureFromLabels()
	for _, v := range groups {
		go v.MainLoop()
	}

	// a caseless select functions as an infinite sleep. Using that here since the group loops are all that really matters from here on
	select {}
}

func configureFromLabels() map[string]LazyGroup {
	// theoretically this could create an issue if people manually hostname their lazytainer instances the same
	// for now the solution is "don't do that"
	// we could do something clever to get around this, but not right now.
	container_id, err := os.Hostname()
	check(err)

	dockerClient, err = client.NewClientWithOpts(client.FromEnv)
	check(err)

	//negotiate API version to prevent "client version is too new" error
	dockerClient.NegotiateAPIVersion(context.Background())

	filter := filters.NewArgs(filters.Arg("id", container_id))
	containers, err := dockerClient.ContainerList(context.Background(), container.ListOptions{All: true, Filters: filter})
	check(err)

	groups := make(map[string]LazyGroup)
	labels := containers[0].Labels

	// iterate through labels, building out config for each group
	prefix := "lazytainer.group."
	for label := range labels {
		if strings.HasPrefix(label, prefix) {
			splitLabelValue := strings.Split(label, ".")
			groupName := splitLabelValue[2]

			// check map to see if group is already created
			_, exists := groups[groupName]
			if exists {
				continue
			}

			// required parameters

			// configure ports
			var ports []uint16
			for _, v := range strings.Split(labels[prefix+groupName+".ports"], ",") {
				val, err := strconv.Atoi(v)
				check(err)
				ports = append(ports, uint16(val))
			}

			// optional parameters

			// configure inactiveTimeout
			inactiveTimeout := uint16(30)
			labelValueAsString, exists := labels[prefix+groupName+".inactiveTimeout"]
			if exists {
				val, err := strconv.Atoi(labelValueAsString)
				check(err)
				inactiveTimeout = uint16(val)
			} else {
				debugLogger.Println("Using default timeout of 60 because " + prefix + groupName + ".inactiveTimeout was not set")
			}

			// configure minPacketThreshold
			minPacketThreshold := uint16(30)
			labelValueAsString, exists = labels[prefix+groupName+".minPacketThreshold"]
			if exists {
				val, err := strconv.Atoi(labelValueAsString)
				check(err)
				minPacketThreshold = uint16(val)
			} else {
				debugLogger.Println("Using default threshold of 30 because " + prefix + groupName + ".minPacketThreshold was not set")
			}

			// configure ignoreActiveClients
			ignoreActiveClients := false
			labelValueAsString, exists = labels[prefix+groupName+".ignoreActiveClients"]
			if exists {
				val, err := strconv.ParseBool(labelValueAsString)
				check(err)
				ignoreActiveClients = val
			} else {
				debugLogger.Println("Using default ignoreActiveClients of false because " + prefix + groupName + ".ignoreActiveClients was not set")
			}

			// configure netInterface
			netInterface := "eth0"
			labelValueAsString, exists = labels[prefix+groupName+".netInterface"]
			if exists {
				netInterface = labelValueAsString
			} else {
				debugLogger.Println("Using default netInterface of eth0 because " + prefix + groupName + ".netInterface was not set")
			}

			// configure pollRate
			pollRate := uint16(30)
			labelValueAsString, exists = labels[prefix+groupName+".pollRate"]
			if exists {
				val, err := strconv.Atoi(labelValueAsString)
				check(err)
				pollRate = uint16(val)
			} else {
				debugLogger.Println("Using default pollRate of 30 because " + prefix + groupName + ".pollRate was not set")
			}

			// configure sleepMethod
			sleepMethod := "stop"
			labelValueAsString, exists = labels[prefix+groupName+".sleepMethod"]
			if exists {
				sleepMethod = labelValueAsString
			} else {
				debugLogger.Println("Using default sleepMethod of stop because " + prefix + groupName + ".sleepMethod was not set")
			}

			groups[groupName] = LazyGroup{
				groupName:           groupName,
				inactiveTimeout:     inactiveTimeout,
				minPacketThreshold:  minPacketThreshold,
				ignoreActiveClients: ignoreActiveClients,
				netInterface:        netInterface,
				pollRate:            pollRate,
				ports:               ports,
				sleepMethod:         sleepMethod,
			}
		}
	}

	for _, g := range groups {
		debugLogger.Printf("%+v\n", g)
	}

	return groups
}

// general error handling
func check(err error) {
	if err != nil {
		// fmt.Println(err)
		panic(err)
	}
}
