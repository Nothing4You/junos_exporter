// SPDX-License-Identifier: MIT

package rpc

import (
	"bytes"
	"encoding/xml"
	"fmt"

	"log"

	"github.com/czerwonk/junos_exporter/pkg/connector"
)

// Parser parses XML of RPC-Output
type Parser func([]byte) error

type ClientOption func(*Client)

func WithDebug() ClientOption {
	return func(cl *Client) {
		cl.debug = true
	}
}

func WithSatellite() ClientOption {
	return func(cl *Client) {
		cl.satellite = true
	}
}

func WithNetconf() ClientOption {
	return func(cl *Client) {
		cl.netconf = true
	}
}

// Client sends commands to JunOS and parses results
type Client struct {
	conn      *connector.SSHConnection
	debug     bool
	satellite bool
	netconf   bool
}

// NewClient creates a new client to connect to
func NewClient(ssh *connector.SSHConnection, opts ...ClientOption) *Client {
	cl := &Client{conn: ssh}

	for _, opt := range opts {
		opt(cl)
	}

	return cl
}

// RunCommandAndParse runs a command on JunOS and unmarshals the XML result
func (c *Client) RunCommandAndParse(cmd string, obj interface{}) error {
	if c.netconf {
		return c.RunCommandAndParseWithParser(cmd, func(b []byte) error {
			//in junos the xml interfaces contains line breaks in the values
			return xml.Unmarshal(bytes.ReplaceAll(b, []byte("\n"), []byte("")), obj)
		})
	} else {
		return c.RunCommandAndParseWithParser(cmd, func(b []byte) error {
			return xml.Unmarshal(b, obj)
		})
	}
}

type rpcReply struct {
	XMLName   xml.Name  `xml:"rpc-reply"`
	Body      []byte    `xml:",innerxml"`
}

// RunCommandAndParseWithParser runs a command on JunOS and unmarshals the XML result using the specified parser function
func (c *Client) RunCommandAndParseWithParser(cmd string, parser Parser) error {
	if c.debug {
		log.Printf("Running command on %s: %s\n", c.conn.Host(), cmd)
	}

	var err error
	var b []byte

	if c.netconf {
		b, err = c.conn.RunCommand(cmd)
	} else {
		b, err = c.conn.RunCommand(fmt.Sprintf("%s | display xml", cmd))
	}

	if err != nil {
		return err
	}

	if c.debug {
		log.Printf("Output for %s: %s\n", c.conn.Host(), string(b))
	}

	if !c.netconf {
		var reply *rpcReply
		err := xml.Unmarshal(b, &reply)

		if err != nil {
			return err
		}

		b = reply.Body
	}

	err = parser(b)
	return err
}

// Device returns device information for the connected device
func (c *Client) Device() *connector.Device {
	return c.conn.Device()
}

// IsSatelliteEnabled returns if sattelite features are enabled on the device
func (c *Client) IsSatelliteEnabled() bool {
	return c.satellite
}

// IsSatelliteEnabled returns if sattelite features are enabled on the device
func (c *Client) IsNetconfEnabled() bool {
	return c.netconf
}
