package main

import (
	"fmt"
	"net"
	"sync"
	"time"
)

var (
	timeout = 10 * time.Second
)

type resultStatus int

func (r resultStatus) String() string {
	switch r {
	case 0:
		return "In Progress"
	case 200:
		return "Success"
	case 401:
		return "Unauthorized"
	case 100:
		return "Trying"
	case 429:
		return "Too Many Requests"
	case 404:
		return "Not Found"
	case 403:
		return "Forbidden"
	default:
		return "Unknown"
	}
}

type connection struct {
	TCPConn           *net.TCPConn
	UDPConn           *net.UDPConn
	Username          string       `json:"username"`
	Password          string       `json:"password"`
	Uri               string       `json:"uri"`
	Port              int          `json:"port"`
	Protocol          string       `json:"protocol"`
	InviteDest        *string      `json:"invite_dest"` // use for invite request for call
	Result            resultStatus `json:"-"`
	IsTCP             bool         `json:"-"`
	ClientIP          string       `json:"-"`
	ClientPort        int          `json:"-"`
	IsRegisterRequest bool         `json:"-"`
}

type Connection interface {
	UDPDial() error
	TCPDial() error
	Read(b []byte) (int, error)
	Write(b []byte) (int, error)
	Close() error
	GetObj() *connection
}

func (c *connection) GetObj() *connection {
	return c
}

func getClientIP() string {
	var s net.Conn
	s, err = net.Dial("udp", "10.255.255.255:1")
	if err != nil {
		return "127.0.0.1"
	}
	defer func(s net.Conn) {
		err = s.Close()
		if err != nil {
			logChan <- logMsg{
				level: 2,
				msg:   "Failed to close socket connection",
			}
		}
	}(s)
	return s.LocalAddr().(*net.UDPAddr).IP.String()
}

func (c *connection) UDPDial() error {
	var udpLocalAddr *net.UDPAddr
	udpLocalAddr, err = net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", c.ClientIP, c.ClientPort))
	var udpSipAddr *net.UDPAddr
	udpSipAddr, err = net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", c.Uri, c.Port))
	if err != nil {
		logChan <- logMsg{
			level: 2,
			msg:   fmt.Sprintf("Error resolving TCP address: %s", err),
		}
		return err
	}
	logChan <- logMsg{
		level: 1,
		msg:   fmt.Sprintf("Server Socket UDP Address: %s", udpSipAddr),
	}
	var conn *net.UDPConn
	conn, err = net.DialUDP("udp", udpLocalAddr, udpSipAddr)
	if err != nil {
		logChan <- logMsg{
			level: 2,
			msg:   fmt.Sprintf("Error binding to UDP address: %s", err),
		}
		return err
	}
	c.ClientPort = conn.LocalAddr().(*net.UDPAddr).Port
	logChan <- logMsg{
		level: 1,
		msg:   fmt.Sprintf("Client Socket UDP Address: %s:%d", c.ClientIP, c.ClientPort),
	}
	err = conn.SetDeadline(time.Now().Add(timeout))
	if err != nil {
		logChan <- logMsg{
			level: 2,
			msg:   fmt.Sprintf("Error setting UDP connection timeout: %s", err),
		}
		return err
	}
	c.IsTCP = false
	c.UDPConn = conn
	return nil
}

func (c *connection) TCPDial() error {
	var tcpLocalAddr *net.TCPAddr
	tcpLocalAddr, err = net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", c.ClientIP, c.ClientPort))
	var tcpSipAddr *net.TCPAddr
	tcpSipAddr, err = net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", c.Uri, c.Port))
	if err != nil {
		logChan <- logMsg{
			level: 2,
			msg:   fmt.Sprintf("Error resolving TCP address: %s", err),
		}
		return err
	}
	logChan <- logMsg{
		level: 1,
		msg:   fmt.Sprintf("Server Socket TCP Address: %s", tcpSipAddr),
	}
	var conn *net.TCPConn
	conn, err = net.DialTCP("tcp", tcpLocalAddr, tcpSipAddr)
	if err != nil {
		logChan <- logMsg{
			level: 2,
			msg:   fmt.Sprintf("Error binding to TCP address: %s", err),
		}
		return err
	}
	c.ClientPort = conn.LocalAddr().(*net.TCPAddr).Port
	logChan <- logMsg{
		level: 1,
		msg:   fmt.Sprintf("Client Socket TCP Address: %s:%d", c.ClientIP, c.ClientPort),
	}
	err = conn.SetDeadline(time.Now().Add(timeout))
	if err != nil {
		logChan <- logMsg{
			level: 2,
			msg:   fmt.Sprintf("Error setting UDP connection timeout: %s", err),
		}
		return err
	}
	c.IsTCP = true
	c.TCPConn = conn

	return nil
}

func (c *connection) Read(b []byte) (int, error) {
	if c.IsTCP {
		return c.TCPConn.Read(b)
	}
	return c.UDPConn.Read(b)
}

func (c *connection) Write(b []byte) (int, error) {
	if c.IsTCP {
		return c.TCPConn.Write(b)
	}
	return c.UDPConn.Write(b)
}

func (c *connection) Close() error {
	if c.IsTCP {
		return c.TCPConn.Close()
	}
	return c.UDPConn.Close()
}

func validateConnections(conn ConnectionList) <-chan Connection {
	var wg sync.WaitGroup
	out := make(chan Connection)

	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			for _, c := range conn {
				obj := c.GetObj()
				obj.ClientIP = getClientIP()
				obj.ClientPort = 0
				if obj.InviteDest == nil {
					obj.IsRegisterRequest = true
				} else {
					obj.IsRegisterRequest = false
				}
				if obj.Protocol == "udp" {
					obj.IsTCP = false
					err = c.UDPDial()
					if err != nil {
						continue
					}
				} else {
					obj.Protocol = "tcp"
					obj.IsTCP = true
					err = c.TCPDial()
					if err != nil {
						continue
					}
				}
				out <- c
			}
			wg.Done()
		}()
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}
