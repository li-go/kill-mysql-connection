package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/go-sql-driver/mysql"
)

var (
	configFile string
	maxTime    int
)

func main() {
	flag.StringVar(&configFile, "config", "", "config file in toml format")
	flag.IntVar(&maxTime, "max-time", 100, "kill process lives longer than max-time seconds")
	flag.Parse()

	if configFile == "" {
		flag.Usage()
		os.Exit(2)
	}

	config, err := readConfig()
	if err != nil {
		panic(err)
	}

	network := "tcp"
	if config.SSHTunnel.UseTunnel {
		network = "ssh+tcp"

		sshClient, err := OpenSSHClient(config.SSHTunnel)
		if err != nil {
			panic(err)
		}
		defer sshClient.Close()

		mysql.RegisterDial(network, func(addr string) (conn net.Conn, e error) {
			return sshClient.Dial("tcp", addr)
		})
	}

	dbClient, err := NewDBClient(config.Mysql, network)
	if err != nil {
		panic(err)
	}
	defer dbClient.Close()

	processlist, err := dbClient.GetProcesslist()
	if err != nil {
		panic(err)
	}
	log.Printf("Got %d processes", len(processlist))

	PrintProcesslist(processlist)

	var killProcesses []Process
	for _, p := range processlist {
		if p.Time >= maxTime {
			killProcesses = append(killProcesses, p)
		}
	}

	fmt.Printf("\nGoing to kill %d processes? Yes/No → ", len(killProcesses))
	ans, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil || strings.ToLower(strings.TrimSpace(ans)) != "yes" {
		return
	}

	var nKilled int
	for _, p := range killProcesses {
		err := dbClient.KillProcess(p.ID)
		if err != nil {
			log.Print(err)
			continue
		}
		nKilled++
	}
	log.Printf("Killed %d processes", nKilled)
}

func readConfig() (*Config, error) {
	config := &Config{}
	_, err := toml.DecodeFile(configFile, config)
	if err != nil {
		return nil, fmt.Errorf("fail to read config: %w", err)
	}
	return config, nil
}

type Config struct {
	Mysql     DBConfig
	SSHTunnel SSHConfig `toml:"ssh_tunnel"`
}
