package main

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"time"
)

var (
	timeout = 10 * time.Second
)

type Status int

func (r Status) String() string {
	switch r {
	case 0:
		return "In Progress"
	case 1:
		return "100 Trying"
	case 2:
		return "200 Success"
	case 3:
		return "401 Unauthorized"

	case 4:
		return "429 Too Many Requests"
	case 5:
		return "404 Not Found"
	case 6:
		return "403 Forbidden"
	case 20:
		return "20 Failed To Open Connection"
	case 21:
		return "21 Preparing Request"
	case 23:
		return "23 Failed"
	default:
		return "Unknown"
	}
}

type Connection struct {
	TCPConn    *net.TCPConn
	UDPConn    *net.UDPConn
	Username   string `json:"username"`
	Password   string `json:"password"`
	IsTCP      bool   `json:"-"`
	Status     Status `json:"-"`
	ClientPort int    `json:"-"`
}

func (c *Connection) sendRequestToServer(header string) error {
	logChan <- logMsg{
		level: 1,
		msg:   fmt.Sprintf("retrying cSeq(%d) ...", cSeq),
	}
	header += "\r\n\r\n"
	logChan <- logMsg{
		level: 1,
		msg:   "Sent Packet Header: \n\t" + strings.Replace(header, "\r\n", "\n\t", -1),
	}
	_, err = c.Write([]byte(header))
	if err != nil {
		return err
	}

	buffer := make([]byte, 1024)
	var response int
	response, err = c.Read(buffer)
	if err != nil {
		logChan <- logMsg{
			level: 3,
			msg:   fmt.Sprintf("Error reading from response: %s", err),
		}
		return err
	}
	sipResponseString := string(buffer[:response])

	logChan <- logMsg{
		level: 1,
		msg:   "Response from server: \n\t" + strings.Replace(sipResponseString, "\r\n", "\n\t", -1),
	}

	if strings.Contains(string(buffer[:response]), "200 OK") {
		c.setResult(2)
		return nil
	}

	if strings.Contains(string(buffer[:response]), "100 Trying") {
		logChan <- logMsg{
			level: 1,
			msg:   "100 Trying",
		}
		c.setResult(1)
		return nil
	}

	if strings.Contains(string(buffer[:response]), "401 Unauthorized") {
		logChan <- logMsg{
			level: 1,
			msg:   "401 Unauthorized",
		}
		nonceFinder(sipResponseString)
		qopFinder(sipResponseString)
		realmFinder(sipResponseString)
		c.setResult(3)
		return nil
	}
	return errors.New("request Failed")
}

func (c *Connection) getObj() Connection {
	return *c
}

func (c *Connection) setResult(i int) {
	c.Status = Status(i)
}

func (c *Connection) getResult() (Status, string) {
	return c.Status, c.Status.String()
}

func (c *Connection) getUsername() string {
	return c.Username
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
				level: 3,
				msg:   "Failed to close socket connection",
			}
		}
	}(s)
	return s.LocalAddr().(*net.UDPAddr).IP.String()
}

func (c *Connection) UDPDial() error {
	var udpLocalAddr *net.UDPAddr
	udpLocalAddr, err = net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", clientIP, 0))
	var udpSipAddr *net.UDPAddr
	udpSipAddr, err = net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", flags.uri, flags.port))
	if err != nil {
		logChan <- logMsg{
			level: 3,
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
			level: 3,
			msg:   fmt.Sprintf("Error binding to UDP address: %s", err),
		}
		return err
	}
	c.ClientPort = conn.LocalAddr().(*net.UDPAddr).Port
	logChan <- logMsg{
		level: 1,
		msg:   fmt.Sprintf("Client Socket UDP Address: %s:%d", clientIP, c.ClientPort),
	}
	err = conn.SetDeadline(time.Now().Add(timeout))
	if err != nil {
		logChan <- logMsg{
			level: 3,
			msg:   fmt.Sprintf("Error setting UDP connection timeout: %s", err),
		}
		return err
	}
	c.IsTCP = false
	c.UDPConn = conn
	return nil
}

func (c *Connection) TCPDial() error {
	var tcpLocalAddr *net.TCPAddr
	tcpLocalAddr, err = net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", clientIP, 0))
	var tcpSipAddr *net.TCPAddr
	tcpSipAddr, err = net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", flags.uri, flags.port))
	if err != nil {
		logChan <- logMsg{
			level: 3,
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
			level: 3,
			msg:   fmt.Sprintf("Error binding to TCP address: %s", err),
		}
		return err
	}
	c.ClientPort = conn.LocalAddr().(*net.TCPAddr).Port
	logChan <- logMsg{
		level: 1,
		msg:   fmt.Sprintf("Client Socket TCP Address: %s:%d", clientIP, c.ClientPort),
	}
	err = conn.SetDeadline(time.Now().Add(timeout))
	if err != nil {
		logChan <- logMsg{
			level: 3,
			msg:   fmt.Sprintf("Error setting UDP connection timeout: %s", err),
		}
		return err
	}
	c.IsTCP = true
	c.TCPConn = conn
	return nil
}

func (c *Connection) Read(b []byte) (int, error) {
	if c.IsTCP {
		return c.TCPConn.Read(b)
	}
	return c.UDPConn.Read(b)
}

func (c *Connection) Write(b []byte) (int, error) {
	if c.IsTCP {
		return c.TCPConn.Write(b)
	}
	return c.UDPConn.Write(b)
}

func (c *Connection) Close() error {
	if c.IsTCP {
		return c.TCPConn.Close()
	}
	return c.UDPConn.Close()
}

func validateConnectionCh(conn Connection) <-chan Connection {
	out := make(chan Connection)
	go func() {
		err = nil
		if flags.protocol == "udp" {
			conn.IsTCP = false
			err = conn.UDPDial()
		} else {
			conn.IsTCP = true
			err = conn.TCPDial()
		}
		if err != nil {
			conn.Status = 20
		}
		out <- conn
		close(out)
	}()
	return out
}
