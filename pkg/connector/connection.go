// SPDX-License-Identifier: MIT

package connector

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/pkg/errors"

	"golang.org/x/crypto/ssh"

	"github.com/nemith/netconf"
	ncssh "github.com/nemith/netconf/transport/ssh"
)

// SSHConnection encapsulates the connection to the device
type SSHConnection struct {
	device   *Device
	client   *ssh.Client
	conn     net.Conn
	lastUsed time.Time
	mu       sync.Mutex
	done     chan struct{}
	netconf bool
	netconfsession *netconf.Session
}

// RunCommand runs a command against the device
func (c *SSHConnection) RunCommand(cmd string) ([]byte, error) {
	if c.netconf	{
		return c.RunCommandNetconf(cmd)
	} else {
		return c.RunCommandSSH(cmd)
	}
}

func (c *SSHConnection) RunCommandSSH(cmd string) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.lastUsed = time.Now()

	if c.client == nil {
		return nil, errors.New("not connected")
	}

	session, err := c.client.NewSession()
	if err != nil {
		return nil, errors.Wrap(err, "could not open session")
	}
	defer session.Close()

	var b = &bytes.Buffer{}
	session.Stdout = b

	err = session.Run(cmd)
	if err != nil {
		return nil, errors.Wrap(err, "could not run command")
	}

	return b.Bytes(), nil
}

func (c *SSHConnection) RunCommandNetconf(cmd string) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	var err error

	if c.client == nil {
		return nil, errors.New("not connected")
	}

	if c.netconfsession == nil {
		t, err := ncssh.NewTransport(c.client)

		if err != nil {
			return nil, errors.Wrap(err, "could not create netconf transport")
		}

		c.netconfsession, err = netconf.Open(t)

		if err != nil {
			return nil, errors.Wrap(err, "could not open netconf session")
		}
	}

	msg := &netconf.RPCMsg{
		Operation: cmd,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	reply, err := c.netconfsession.Do(ctx, msg)

	if err != nil {
		if err == io.EOF {
			//probably lost the session, closing to force a reopen
			fmt.Println("Error - Closing")
			c.netconfsession.Close(ctx)
			c.netconfsession = nil
		}
		return nil, errors.Wrap(err, "could not run command")
	}

	return reply.Body, nil
}

func (c *SSHConnection) isConnected() bool {
	return c.conn != nil
}

func (c *SSHConnection) terminate() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.conn.Close()

	c.client = nil
	c.conn = nil
}

func (c *SSHConnection) close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.client != nil {
		c.client.Close()
	}

	c.done <- struct{}{}
	c.conn = nil
	c.client = nil
}

// Host returns the hostname of the connected device
func (c *SSHConnection) Host() string {
	return c.device.Host
}

// Device returns the device information of the connected device
func (c *SSHConnection) Device() *Device {
	return c.device
}
