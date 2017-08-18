package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"time"

	"bashdrew/demochatserver/lobby"
	chatlog "bashdrew/demochatserver/logger"
)

const (
	CONN_HOST = "localhost"
	CONN_PORT = "8085"
	CONN_TYPE = "tcp"
	ROOMS_MAX = 10
	LOG_FILE  = "chattelnet.log"
)

type Config struct {
	Server `json:"server"`
	Lobby  `json:"lobby"`
	Log    `json:"log"`
}

type Server struct {
	Host string `json:"host"`
	Port string `json:"port"`
	Type string `json:"type"`
}

type Lobby struct {
	RoomsMax int `json:"roomsmax"`
}

type Log struct {
	FileName string `json:"filename"`
}

func LoadConfiguration(file string) Config {
	var config Config

	configFile, err := os.Open(file)
	defer configFile.Close()

	if err != nil {
		log.Fatal(err.Error())
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)

	return config
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	chatConfig := Config{
		Server: Server{Host: CONN_HOST, Port: CONN_PORT, Type: CONN_TYPE},
		Lobby:  Lobby{RoomsMax: ROOMS_MAX},
		Log:    Log{FileName: LOG_FILE},
	}

	if len(os.Args) == 2 {
		chatConfig = LoadConfiguration(os.Args[1])
	}

	logFile := chatlog.InitializeLogFile(chatConfig.Log.FileName)
	defer logFile.Close()

	//	mainLobby := lobby.NewLobby(chatConfig.Lobby.RoomsMax, logFile)
	mainLobby := lobby.NewLobby(chatConfig.Lobby.RoomsMax)

	listener, err := net.Listen(chatConfig.Server.Type, chatConfig.Server.Host+":"+chatConfig.Server.Port)
	if err != nil {
		chatlog.PrintLog(fmt.Sprintln("Error listening:", err.Error()))
		os.Exit(1)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			chatlog.PrintLog(fmt.Sprintf("Error in Accept(): %s\n", err.Error()))
			os.Exit(1)
		}

		mainLobby.AcceptConnection(conn)
	}
}
