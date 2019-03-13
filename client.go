package osc

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"time"
)

// Client enables you to send OSC packets. It sends OSC messages and bundles to
// the given IP address and port.
type Client struct {
	ip         string
	port       int
	laddr      *net.UDPAddr
	connection *net.UDPConn
}

// NewClient creates a new OSC client. The Client is used to send OSC
// messages and OSC bundles over an UDP network connection. The `ip` argument
// specifies the IP address and `port` defines the target port where the
// messages and bundles will be send to.
func NewClient(ip string, port int) *Client {
	return &Client{ip: ip, port: port, laddr: nil}
}

func (c *Client) Connect(retries int) error {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", c.ip, c.port))
	if err != nil {
		return err
	}

	for c.connection == nil && retries > 0 {
		if c.connection, err = net.DialUDP("udp", nil, addr); err != nil {
			time.Sleep(2 * time.Second)
		}
		retries--
	}

	if c.connection == nil {
		return fmt.Errorf("Unable to connect to %v:%v.", c.ip, c.port)
	}

	return nil
}

func (c *Client) Disconnect() error {
	return c.connection.Close()
}

// SetLocalAddr sets the local address.
func (c *Client) SetLocalAddr(ip string, port int) error {
	laddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		return err
	}
	c.laddr = laddr
	return nil
}

// Send sends an OSC Bundle or an OSC Message.
func (c *Client) Send(message *Message) error {
	if c.connection == nil {
		return fmt.Errorf("Unable to send, not connected.")
	}

	data, err := message.MarshalBinary()
	if err != nil {
		return err
	}

	if _, err = c.connection.Write(data); err != nil {
		return err
	}
	return nil
}

// Receive listens for messages (replies, this is not a server) until the timeout is expired
func (c *Client) Receive(timeout time.Duration) (error, *Message) {
	buffer := make([]byte, 65535)
	c.connection.SetReadDeadline(time.Now().Add(timeout))
	_, _, err := c.connection.ReadFrom(buffer)
	if err != nil {
		return err, nil
	}

	var start int
	msg, err := readMessage(bufio.NewReader(bytes.NewBuffer(buffer)), &start)
	if err != nil {
		return err, nil
	}

	return nil, msg
}
