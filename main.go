package main

import (
	"net"
	"os"
	"path/filepath"
	"sync"

	"tunnel/tunnel"

	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
)

var wg sync.WaitGroup

func main() {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	viper.AddConfigPath(dir)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
	sshConfig := &ssh.ClientConfig{
		User: viper.GetString("ssh.username"),
		Auth: []ssh.AuthMethod{
			ssh.Password(viper.GetString("ssh.password")),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	ep := []*tunnel.EndpointPair{}
	arr := viper.Get("tunnel").([]interface{})
	for _, v := range arr {
		vv := v.(map[interface{}]interface{})
		localPort := vv["localport"].(int)
		remotePort := vv["remoteport"].(int)
		ep = append(ep, &tunnel.EndpointPair{Local: &tunnel.Endpoint{Host: "localhost", Port: localPort},
			Remote: &tunnel.Endpoint{Host: "0.0.0.0", Port: remotePort}})
	}

	tun := &tunnel.SSHtunnel{
		Config: sshConfig,
		Server: &tunnel.Endpoint{
			Host: viper.GetString("ssh.server"),
			Port: viper.GetInt("ssh.port"),
		},
		Pairs: ep,
	}

	tun.Start(&wg)
	wg.Wait()
}
