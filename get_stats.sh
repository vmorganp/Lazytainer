#!/bin/sh
while [ 1 ];
do 
    echo "$(netstat -n | grep $PORT | wc -l) active connections on $PORT"
    echo "tx packets: $(cat /sys/class/net/eth0/statistics/tx_packets)"
    echo "rx packets: $(cat /sys/class/net/eth0/statistics/rx_packets)"
    sleep 1;
done