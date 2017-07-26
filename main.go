package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"net"
	"golang.org/x/net/ipv4"
	"github.com/songgao/water"
)


const (
	BUFFERSIZE = 1500
	MTU = "1300"
)

var (
	localIP = flag.String("local", "", "Local run interface IP like 192.168.3.3")
	remoteIP = flag.String("remote", "", "Remote run interface IP like 192.168.4.3")
	sendQueue = flag.String("sendqueue", "", "AWS SQS queue to send to")
	receiveQueue = flag.String("receivequeue", "", "AWS SQS queue to receive from")
)

func runIP(args ...string) {
	cmd := exec.Command("/sbin/ip", args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	err := cmd.Run()
	if nil != err {
		log.Fatal("Error running /sbin/ip:", err)
	}
}

func main() {
	flag.Parse()

	if "" == *localIP {
		flag.Usage()
		log.Fatalln("\nlocal ip is not specified")
	}

	if "" == *remoteIP {
		flag.Usage()
		log.Fatalln("\nremote ip is not specified")
	}

	if "" == *sendQueue {
		flag.Usage()
		log.Fatalln("\nsendqueue is not specified")
	}

	if "" == *receiveQueue {
		flag.Usage()
		log.Fatalln("\nreceivequeue is not specified")
	}


	sender := newQ(*sendQueue)
	receiver := newQ(*receiveQueue)

	config := water.Config{
		DeviceType: water.TUN,
	}

	iface, err := water.New(config)
	if nil != err {
		log.Fatalln("Unable to allocate TUN interface:", err)
	}
	log.Println("Interface allocated:", iface.Name())

	runIP("link", "set", "dev", iface.Name(), "mtu", MTU)
	runIP("addr", "add", *localIP, "dev", iface.Name())
	runIP("link", "set", "dev", iface.Name(), "up")
	runIP("route", "add", *remoteIP +"/32", "dev", iface.Name())

	remoteIPAddr := net.ParseIP(*remoteIP)
	go func() {
		for {
			received, msgs, err := receiver.receive()
			if err != nil {
				fmt.Println("Error receiving:", err)
				continue
			}

			if received {
				for _, msg := range msgs {
					n := len(msg)
					iface.Write(msg[:n])
				}
			}
		}
	}()

	packet := make([]byte, BUFFERSIZE)
	for {
		plen, err := iface.Read(packet)
		if err != nil {
			break
		}
		header, _ := ipv4.ParseHeader(packet[:plen])
		if remoteIPAddr.Equal(header.Dst) {
			sender.send(packet[:plen])
		}
	}
}
