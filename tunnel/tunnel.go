package tunnel

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

var sshClient *ssh.Client
var wg *sync.WaitGroup
var connected bool = false
var connecting bool = false

//Endpoint struct
type Endpoint struct {
	Host string
	Port int
}

//EndpointPair struct
type EndpointPair struct {
	Local  *Endpoint
	Remote *Endpoint
}

func (endpoint *Endpoint) String() string {
	return fmt.Sprintf("%s:%d", endpoint.Host, endpoint.Port)
}

//SSHtunnel struct
type SSHtunnel struct {
	Server *Endpoint
	Config *ssh.ClientConfig

	Pairs []*EndpointPair
}

//Start function
func (tun *SSHtunnel) Start(wg *sync.WaitGroup) {
	wg.Add(1)
	time.Sleep(2 * time.Second)
	connecting = true
	tun.start()
	for {
		select {
		case <-time.After(1 * time.Minute):
			if !connected && !connecting {
				connecting = true
				tun.start()
			}
		}
	}
}

func (tun *SSHtunnel) start() error {
	var err error
	fmt.Println("LOGGIN TO SERVER...")
	sshClient, err = ssh.Dial("tcp", tun.Server.String(), tun.Config)
	if err != nil {
		fmt.Println("LOGGIN ERROR!")
		fmt.Println(err)
		connecting = false
		connected = false
		return err
	}
	connected = true
	connecting = false
	fmt.Println("LOGGED successfully!")

	for _, item := range tun.Pairs {
		listener, err := sshClient.Listen("tcp", item.Remote.String())
		if err != nil {
			fmt.Println(err)
			sshClient.Close()
			sshClient.Wait()
			connecting = false
			connected = false
			time.Sleep(5 * time.Minute)
			return err
		}
		fmt.Println("LISTENING on remote ", item.Remote.String())

		go func(l net.Listener, e *Endpoint) {
			for {
				conn, err := l.Accept()
				if err != nil {
					return
				}
				go tun.forward(conn, e)
			}
		}(listener, item.Local)
	}

	go func() {
		sshClient.Wait()
		connected = false
		connecting = false
	}()

	return nil
}

func (tun *SSHtunnel) forward(sshCon net.Conn, local *Endpoint) {
	localCon, err := net.Dial("tcp", local.String())
	if err != nil {
		fmt.Println("Error dial local ", err)
		return
	}

	copyConn := func(writer, reader net.Conn) {
		defer writer.Close()
		defer reader.Close()

		_, err := io.Copy(writer, reader)
		if err != nil {
			fmt.Printf("io.Copy error: %s", err)
		}
	}

	go copyConn(sshCon, localCon)
	go copyConn(localCon, sshCon)
}
