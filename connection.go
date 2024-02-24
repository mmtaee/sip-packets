package main

import (
	"fmt"
	"net"
	"time"
)

var (
	clientIP   = getClientIP()
	clientPort = 0
	timeout    = 10 * time.Second
)

type connection struct {
	tcpConn *net.TCPConn
	udpConn *net.UDPConn
	isTCP   bool
}

type Connection interface {
	udpDial() error
	tcpDial() error
	Read(b []byte) (int, error)
	Write(b []byte) (int, error)
	Close() error
}

func getClientIP() string {
	var s net.Conn
	s, err = net.Dial("udp", "10.255.255.255:1")
	if err != nil {
		//fmt.Println("Error in get Client IP address:", err)
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

func (c *connection) udpDial() error {
	var udpLocalAddr *net.UDPAddr
	udpLocalAddr, err = net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", clientIP, clientPort))
	var udpSipAddr *net.UDPAddr
	udpSipAddr, err = net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", sipUri, sipPort))
	if err != nil {
		logChan <- logMsg{
			level: 2,
			msg:   fmt.Sprintf("Error resolving TCP address: %s", err),
		}
		return err
	}
	logChan <- logMsg{
		level: 1,
		msg:   fmt.Sprintf("Client Socket UDP Address: %s", udpSipAddr),
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
	clientPort = conn.LocalAddr().(*net.UDPAddr).Port
	err = conn.SetDeadline(time.Now().Add(timeout))
	if err != nil {
		logChan <- logMsg{
			level: 2,
			msg:   fmt.Sprintf("Error setting UDP connection timeout: %s", err),
		}
		return err
	}
	c.isTCP = false
	c.udpConn = conn
	return nil
}

func (c *connection) tcpDial() error {
	var tcpLocalAddr *net.TCPAddr
	tcpLocalAddr, err = net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", clientIP, clientPort))
	var tcpSipAddr *net.TCPAddr
	tcpSipAddr, err = net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", sipUri, sipPort))
	if err != nil {
		logChan <- logMsg{
			level: 2,
			msg:   fmt.Sprintf("Error resolving TCP address: %s", err),
		}
		return err
	}
	logChan <- logMsg{
		level: 1,
		msg:   fmt.Sprintf("Client Socket TCP Address: %s", tcpSipAddr),
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
	clientPort = conn.LocalAddr().(*net.TCPAddr).Port
	err = conn.SetDeadline(time.Now().Add(timeout))
	if err != nil {
		logChan <- logMsg{
			level: 2,
			msg:   fmt.Sprintf("Error setting UDP connection timeout: %s", err),
		}
		return err
	}
	c.isTCP = true
	c.tcpConn = conn
	return nil
}

func (c *connection) Read(b []byte) (int, error) {
	if c.isTCP {
		return c.tcpConn.Read(b)
	}
	return c.udpConn.Read(b)
}

func (c *connection) Write(b []byte) (int, error) {
	if c.isTCP {
		return c.tcpConn.Write(b)
	}
	return c.udpConn.Write(b)
}

func (c *connection) Close() error {
	if c.isTCP {
		return c.tcpConn.Close()
	}
	return c.udpConn.Close()
}
