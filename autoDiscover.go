package osc

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"time"
)

func AutoDiscover(port int, msg *Message) (map[string]*Message, error) {
	// listen for udp packets
	c, err := net.ListenPacket("udp", ":0")
	if err != nil {
		return nil, err
	}
	defer c.Close()

	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	//get all broadcast addresses
	var bcAddresses []net.IP
	for _, i := range ifaces {
		addrs, _ := i.Addrs()
		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				v4IP := v.IP.To4()
				if v4IP != nil && !v.IP.IsLoopback() {
					var bcA [4]byte
					for c := byte(0); c <= 3; c++ {
						bcA[c] = v4IP[c] | ^v4IP.DefaultMask()[c]
					}

					bcAddresses = append(bcAddresses, net.IPv4(bcA[0], bcA[1], bcA[2], bcA[3]))
				}
			}
		}
	}

	//sent out discovery message to all networks connected to this machine
	data, _ := msg.MarshalBinary()

	for _, bcAddr := range bcAddresses {
		dst, _ := net.ResolveUDPAddr("udp", fmt.Sprintf("%v:%v", bcAddr.String(), port))
		c.WriteTo(data, dst)
	}

	//receive responses
	b := make([]byte, 512)
	discoveries := make(map[string]*Message)
	readErr := false

	for !readErr {
		c.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
		_, peer, err := c.ReadFrom(b)
		if err == nil {

			var start int
			inMsg, err := readMessage(bufio.NewReader(bytes.NewBuffer(b)), &start)
			if err == nil {
				discoveries[peer.String()] = inMsg
			}
		}
		if err != nil {
			readErr = true
		}
	}

	return discoveries, nil
}
