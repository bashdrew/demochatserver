package chatroom

import (
	"fmt"
	"math/rand"
	"net"
	"strings"
	"time"

	"bashdrew/demochatserver/chatclient"
	chatlog "bashdrew/demochatserver/logger"
)

const (
	ENTEREDROOM_MSG = "\n%s just entered room no: %d\n"
	ALLOWBLOCK_MSG  = "\n%s wants to chat with you.  Allow (Y/N)? "
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}

	return string(b)
}

type ChatMessage struct {
	client  *chatclient.Client
	message string
}

type ChatRoom struct {
	name      string
	number    int
	clients   map[*chatclient.Client]*chatclient.Client
	chanconn  chan net.Conn
	msgin     chan string
	msgout    chan string
	chatMsgIn chan ChatMessage
}

func (chatRoom *ChatRoom) Broadcast(data string) {
	for _, client := range chatRoom.clients {
		client.OutputMessage(data)
	}
}

func (chatRoom *ChatRoom) BroadcastAdmMsg(data ChatMessage) {
	for _, client := range chatRoom.clients {
		client.OutputChatMessage(data.message, client != data.client)
	}
}

func (chatRoom *ChatRoom) BroadcastMsg(data ChatMessage) {
	timeNowStr := time.Now().Format("2006-01-02 15:04:05.000")

	if data.client.GetStatus() != chatclient.DISCONNECTED_STAT {
		for _, client := range chatRoom.clients {
			if client != data.client {
				clientMsg := fmt.Sprintf("\n[%s] %s: %s", timeNowStr, data.client.GetUserName(), data.message)
				if client.InTheList(data.client) {
					if client.IsAllowed(data.client) {
						client.OutputChatMessage(clientMsg, client != data.client)
					}
				} else {
					client.SetClientToAllow(data.client)
					client.SetClientMsg(clientMsg)
					client.SetMode(chatclient.ALLOWBLOCK_MODE)
					client.OutputChatMessage(fmt.Sprintf(ALLOWBLOCK_MSG, data.client.GetUserName()), false)
				}
			} else {
				client.OutputMessage(fmt.Sprintf("%s: ", "You"))
			}
		}
	}
}

func (chatRoom *ChatRoom) Join(client *chatclient.Client) {
	chatRoom.clients[client] = client
	joinMsg := fmt.Sprintf(ENTEREDROOM_MSG, client.GetUserName(), chatRoom.GetRoomNo())
	chatRoom.BroadcastAdmMsg(ChatMessage{
		client:  client,
		message: joinMsg,
	})
	chatlog.PrintLog(joinMsg)

	client.OutputMessage(fmt.Sprintf("%s: ", "You"))
	go func() {
		for {
			clientStatus := client.GetStatus()
			if clientStatus != chatclient.DISCONNECTED_STAT {
				clientmsg := client.InputMessage()

				chatRoom.ProcessClientMessage(client, clientmsg)
			} else {
				admMsg := fmt.Sprintf("%v left the room.\n", client.GetUserName())
				chatRoom.BroadcastAdmMsg(ChatMessage{
					client:  client,
					message: "\n" + admMsg,
				})
				chatlog.PrintLog(admMsg)
				delete(chatRoom.clients, client)

				break
			}
		}
	}()
}

func (chatRoom *ChatRoom) ProcessClientMessage(client *chatclient.Client, msg string) {
	switch clientMode := client.GetMode(); clientMode {
	case chatclient.CHAT_MODE:
		clientChatMsg := ChatMessage{client: client, message: msg}
		chatRoom.chatMsgIn <- clientChatMsg
	case chatclient.ALLOWBLOCK_MODE:
		targetClient := client.GetClientToAllow()
		if string(strings.ToUpper(msg)[0]) == "Y" {
			client.AddToList(targetClient, true)
			targetClient.AddToList(client, true)
			client.OutputChatMessage(client.GetClientMsg(), true)
			chatlog.PrintLog(fmt.Sprintf("%s allowed %s", client.GetUserName(), targetClient.GetUserName()))
		} else {
			client.AddToList(targetClient, false)
			targetClient.AddToList(client, false)
			client.OutputChatMessage("", true)
			chatlog.PrintLog(fmt.Sprintf("%s blocked %s", client.GetUserName(), targetClient.GetUserName()))
		}
		client.SetMode(chatclient.CHAT_MODE)
	}
}

func (chatRoom *ChatRoom) InputMessage(data string) {
	chatRoom.msgin <- data
}

func (chatRoom *ChatRoom) Listen() {
	go func() {
		for {
			select {
			case data := <-chatRoom.msgin:
				chatRoom.Broadcast(data)
			case chatData := <-chatRoom.chatMsgIn:
				chatlog.PrintLog(fmt.Sprintf("Message from %s: %s", chatData.client.GetUserName(), chatData.message))
				chatRoom.BroadcastMsg(chatData)
			}
		}
	}()
}

func (chatRoom *ChatRoom) AcceptConnection(conn net.Conn) {
	chatRoom.chanconn <- conn
}

func (chatRoom *ChatRoom) GetName() string {
	return chatRoom.name
}

func (chatRoom *ChatRoom) GetRoomNo() int {
	return chatRoom.number
}

func (chatRoom *ChatRoom) SetRoomNo(roomNo int) {
	chatRoom.number = roomNo
}

func NewChatRoom() *ChatRoom {
	chatRoom := &ChatRoom{
		name:      RandStringRunes(10),
		clients:   make(map[*chatclient.Client]*chatclient.Client, 0),
		chanconn:  make(chan net.Conn),
		msgin:     make(chan string),
		msgout:    make(chan string),
		chatMsgIn: make(chan ChatMessage),
	}

	chatRoom.Listen()

	return chatRoom
}
