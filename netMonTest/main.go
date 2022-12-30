package main

import (
	"fmt"
	"time"

	"github.com/google/gopacket"
	_ "github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

func main() {
	go netMonTest()
	for {
		time.Sleep(time.Hour * 1)
	}
	// http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	// 	fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
	// })

	// http.HandleFunc("/hi", func(w http.ResponseWriter, r *http.Request) {
	// 	fmt.Fprintf(w, "Hi")
	// })

	// log.Fatal(http.ListenAndServe(":80", nil))
}

func netMonTest() {
	defaultSnapLen := int32(262144)
	handle, err := pcap.OpenLive("eth0", defaultSnapLen, true,
		pcap.BlockForever)
	if err != nil {
		panic(err)
	}
	defer handle.Close()

	// if err := handle.SetBPFFilter("port 80 or port 81"); err != nil {
	if err := handle.SetBPFFilter("port 80"); err != nil {
		panic(err)
	}

	totalPAks := 0
	packets := gopacket.NewPacketSource(
		handle, handle.LinkType()).Packets()
	totalPAks += len(packets)
	for range packets {
		totalPAks += len(packets) + 1 // I assume +1 because the for pops one?
		for len(packets) > 0 {
			<-packets
		}
		// fmt.Println(len(packets))
		fmt.Println(totalPAks)

		// fmt.Println(pkt.Metadata().CaptureInfo.)
		// 	// // Your analysis here!
	}
}
