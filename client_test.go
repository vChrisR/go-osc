package osc

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	c := NewClient("localhost", 8967, "localhost", 41789)
	expectedAddr := "127.0.0.1:41789"
	if c.laddr.String() != expectedAddr {
		t.Errorf("Expected laddr to be %s but was %s", expectedAddr, c.laddr.String())
	}

	expectedIp := "localhost"
	if c.ip != expectedIp {
		t.Errorf("Expected ip to be %s but was %s", expectedIp, c.ip)
	}
}

func TestConnectDisconnect(t *testing.T) {
	c := NewClient("localhost", 8967, "localhost", 41789)
	err := c.Connect(1)
	if err != nil {
		t.Errorf("Expected err to be nil but got %s", err.Error())
	}

	if err = checkConnection(c); err != nil {
		t.Errorf(err.Error())
	}

	err = c.Disconnect()
	if err != nil {
		t.Errorf("Disconnect: Expected no error returned but got: %s", err)
	}

	f, _ := c.connection.File()
	if f != nil {
		t.Errorf("Expected connection file to be closed after disconnect but it is still open.")
	}
}

func checkConnection(c *client) error {
	if c.connection == nil {
		return fmt.Errorf("Expected connection to be not nil. Got: nil")
	}

	expectedAddr := "127.0.0.1:41789"
	if c.connection.LocalAddr().String() != expectedAddr {
		return fmt.Errorf("Expected localaddrr to be %s but was %s", expectedAddr, c.connection.LocalAddr().String())
	}

	return nil
}

func TestSendReceive(t *testing.T) {
	returnChan := make(chan []byte)

	pc, err := net.ListenPacket("udp", ":8967")
	if err != nil {
		t.Errorf(err.Error())
	}
	defer pc.Close()

	go func() {
		buf := make([]byte, 2048)
		for {
			_, addr, err := pc.ReadFrom(buf)
			if err != nil {
				fmt.Println(err)
				continue
			}
			_, err = pc.WriteTo(buf, addr)
			if err != nil {
				fmt.Println(err)
			}
			returnChan <- buf
			close(returnChan)
			return
		}
	}()

	c := NewClient("localhost", 8967, "", 0)
	c.Connect(1)
	defer c.Disconnect()

	c.Send(NewMessage("/test"))

	//this tests the reply from the server which should be identical to what we send
	err, nmsg := c.Receive(1 * time.Second)
	if err != nil {
		t.Errorf("Receive error should be nil but got: %s", err)
	} else {
		if nmsg.Address != "/test" {
			t.Errorf("message address should be /test but got: %s", nmsg.Address)
		}
	}

	buf := <-returnChan
	//this tests the content of on the "server" side
	var start int = 0
	msg, err := readMessage(bufio.NewReader(bytes.NewBuffer(buf)), &start)
	if err != nil {
		t.Errorf("Expected readMEssage error to be nil but got: %s", err)
	}

	if msg.Address != "/test" {
		t.Errorf("Expected msg Address to be /test but got: %s", msg.Address)
	}

}

func TestSendDisconnected(t *testing.T) {
	c := NewClient("localhost", 8967, "", 0)
	err := c.Send(NewMessage("/test"))
	if err == nil {
		t.Errorf("Expected err to be not nil but got nil")
	}
}

func TestReceiveFail(t *testing.T) {
	c := NewClient("localhost", 8967, "", 0)
	err, _ := c.Receive(1 * time.Second)
	if err == nil {
		t.Errorf("Expected err to be not nil but got nil")
	}

	c.Connect(1)
	err, _ = c.Receive(1 * time.Second)
	if err == nil {
		t.Errorf("Expected err to be not nil but got nil")
	}
}
