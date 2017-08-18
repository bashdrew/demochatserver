package lobby

import (
	"fmt"
	"net"
	"strconv"

	"bashdrew/demochatserver/chatclient"
	"bashdrew/demochatserver/chatroom"
	chatlog "bashdrew/demochatserver/logger"
)

const (
	WELCOME_MSG      = "Welcome to my Demo Chat Server.\n"
	ENTERNAME_MSG    = "Please enter your username: "
	ROOMLISTHDR_MSG  = "Available Rooms \n"
	ROOMSELECT_MSG   = "Please enter a room number: "
	ROOMNOTFOUND_MSG = "Room no. not found: %s\n"
)

type Lobby struct {
	chatRooms []*chatroom.ChatRoom
	clients   map[*chatclient.Client]*chatclient.Client
	chanconn  chan net.Conn
	msgin     chan string
	msgout    chan string
}

//func NewLobby(maxRooms int, logFile *os.File) *Lobby {
func NewLobby(maxRooms int) *Lobby {
	lobby := &Lobby{
		chatRooms: make([]*chatroom.ChatRoom, 0),
		clients:   make(map[*chatclient.Client]*chatclient.Client, 0),
		chanconn:  make(chan net.Conn),
		msgin:     make(chan string),
		msgout:    make(chan string),
	}

	for i := 0; i < maxRooms; i++ {
		lobby.chatRooms = append(lobby.chatRooms, chatroom.NewChatRoom())
		lobby.chatRooms[i].SetRoomNo(i)
	}

	lobby.Listen()

	return lobby
}

func (lobby *Lobby) Listen() {
	go func() {
		for {
			select {
			case data := <-lobby.msgin:
				lobby.Broadcast(data)
			case conn := <-lobby.chanconn:
				lobby.Join(conn)
			}
		}
	}()
}

func (lobby *Lobby) AcceptConnection(conn net.Conn) {
	lobby.chanconn <- conn
}

func (lobby *Lobby) Join(connection net.Conn) {
	client := chatclient.NewClient(connection)

	client.OutputMessage(WELCOME_MSG)
	client.OutputMessage(ENTERNAME_MSG)
	client.SetStatus(chatclient.WAITING_STAT)
	client.SetMode(chatclient.LOGIN_MODE)
	lobby.clients[client] = client

	go func() {
		for {
			clientStatus := client.GetStatus()
			if clientStatus != chatclient.DISCONNECTED_STAT {
				switch clientStatus {
				case chatclient.WAITING_STAT:
					clientmsg := client.InputMessage()
					lobby.ProcessClientMessage(client, clientmsg)
				}
			} else {
				delete(lobby.clients, client)
				chatlog.PrintLog(fmt.Sprintf("%v disconnected.\n", client.GetUserName()))
				connection.Close()

				break
			}
		}
	}()
}

func (lobby *Lobby) ProcessClientMessage(client *chatclient.Client, msg string) {
	msg = chatlog.TrimInput(msg)

	switch clientMode := client.GetMode(); clientMode {
	case chatclient.LOGIN_MODE:
		client.SetUserName(msg)

		client.SetMode(chatclient.ROOMNO_MODE)
		lobby.displayRoomMsg(client)
	case chatclient.ROOMNO_MODE:
		if roomNo := lobby.EnterChatRoom(msg); roomNo != -1 {
			client.SetRoomNo(roomNo)
			client.SetStatus(chatclient.IN_STAT)
			client.SetMode(chatclient.CHAT_MODE)
			lobby.chatRooms[roomNo].Join(client)
		} else {
			client.OutputMessage(fmt.Sprintf(ROOMNOTFOUND_MSG, msg))
			lobby.displayRoomMsg(client)
		}
	}
}

func (lobby *Lobby) Broadcast(data string) {
	for _, client := range lobby.clients {
		client.OutputMessage(data)
	}
}

func (lobby *Lobby) displayRoomMsg(client *chatclient.Client) {
	roomList := lobby.GetRoomsList()
	client.OutputMessage(roomList)
	client.OutputMessage(ROOMSELECT_MSG)
}

func (lobby *Lobby) GetRoomsList() (result string) {
	result = ROOMLISTHDR_MSG
	for idx, chatRoom := range lobby.chatRooms {
		dtl := fmt.Sprintf("\t[%d] - %s\n", idx, chatRoom.GetName())
		result = result + dtl
	}

	return
}

func (lobby *Lobby) EnterChatRoom(roomNoStr string) (result int) {
	result = -1

	if roomNo, err := strconv.Atoi(roomNoStr); err == nil {
		if roomNo >= 0 && roomNo < len(lobby.chatRooms) {
			result = roomNo
		}
	}
	return
}
