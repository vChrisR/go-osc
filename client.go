package osc

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"time"
)

// osc Client exposes Connect, Disconnect, Send Receive.
type Client interface {
	Connect(int) error
	Disconnect() error
	Send(*Message) error
	Receive(time.Duration) (error, *Message)
}

// Client enables you to send OSC packets. It sends OSC messages and bundles to
// the given IP address and port.
type client struct {
	ip         string
	port       int
	laddr      *net.UDPAddr
	connection *net.UDPConn
	buffer     []byte //use one receive buffer instead of creating a new one with each receive(). This dramatically reduces GC.
}

// NewClient creates a new OSC client. The Client is used to send OSC
// messages and OSC bundles over an UDP network connection. The `ip` argument
// specifies the IP address and `port` defines the target port where the
// messages and bundles will be send to.
func NewClient(ip string, port int, localIP string, localPort int) Client {
	c := &client{
		ip:     ip,
		port:   port,
		laddr:  nil,
		buffer: make([]byte, 1024),
	}

	if localIP != "" {
		c.laddr, _ = net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", localIP, localPort))
	}

	return c
}

// Connect Creates a net.UDPConn. This connection stays open until Disconnect is called
func (c *client) Connect(retries int) error {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", c.ip, c.port))
	if err != nil {
		return err
	}

	for c.connection == nil && retries > 0 {
		if c.connection, err = net.DialUDP("udp", c.laddr, addr); err != nil {
			time.Sleep(2 * time.Second)
		}
		retries--
	}

	if c.connection == nil {
		return fmt.Errorf("Unable to connect to %v:%v.", c.ip, c.port)
	}

	return nil
}

// Disconnect closes the opened net.UDPConn
func (c *client) Disconnect() error {
	return c.connection.Close()
}

// Send sends an OSC Message.
func (c *client) Send(message *Message) error {
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
func (c *client) Receive(timeout time.Duration) (error, *Message) {
	if c.connection == nil {
		return fmt.Errorf("Unable to receive, not connected."), nil
	}

	c.connection.SetReadDeadline(time.Now().Add(timeout))
	_, _, err := c.connection.ReadFrom(c.buffer)
	if err != nil {
		return err, nil
	}

	var start int
	msg, err := readMessage(bufio.NewReader(bytes.NewBuffer(c.buffer)), &start)
	if err != nil {
		return err, nil
	}

	return nil, msg
}
