package chatclient

import (
	"bufio"
	"fmt"
	"io"
	"net"
)

const (
	WAITING_STAT      = 1
	IN_STAT           = 2
	OUT_STAT          = 3
	DISCONNECTED_STAT = 4
	LOGIN_MODE        = 1
	ROOMNO_MODE       = 2
	CHAT_MODE         = 3
	ALLOWBLOCK_MODE   = 4
)

type Client struct {
	status        int
	mode          int
	roomNo        int
	userName      string
	msgin         chan string
	msgout        chan string
	msgAdmOut     chan string
	allowList     map[*Client]bool
	clientToAllow *Client
	clientMsg     string
	reader        *bufio.Reader
	writer        *bufio.Writer
}

func (cc *Client) Receive() {
	for {
		var line string
		var err error

		if line, err = cc.reader.ReadString('\n'); err == io.EOF {
			cc.status = DISCONNECTED_STAT
		} else {
			cc.msgin <- line
		}
	}
}

func (cc *Client) Send() {
	/*	for data := range cc.msgAdmOut {
			cc.writer.WriteString(data)
			cc.writer.Flush()
		}
	*/

	for data := range cc.msgout {
		cc.writer.WriteString(data)
		cc.writer.Flush()
	}
}

func (cc *Client) InputMessage() (data string) {
	data = <-cc.msgin

	return
}

func (cc *Client) OutputMessage(data string) {
	cc.msgout <- data
}

func (cc *Client) OutputChatMessage(data string, displayPrompt bool) {
	cc.msgout <- data

	if displayPrompt {
		cc.msgout <- fmt.Sprintf("%s: ", "You")
	}
}

func (cc *Client) OutputAdmMessage(data string, displayPrompt bool) {
	cc.msgAdmOut <- data

	if displayPrompt {
		cc.msgout <- fmt.Sprintf("%s: ", "You")
	}
}

func (cc *Client) Listen() {
	go cc.Receive()
	go cc.Send()
}

func (cc *Client) GetStatus() int {
	return cc.status
}

func (cc *Client) SetStatus(stat int) {
	cc.status = stat
}

func (cc *Client) GetMode() int {
	return cc.mode
}

func (cc *Client) SetMode(mode int) {
	cc.mode = mode
}

func (cc *Client) GetRoomno() int {
	return cc.roomNo
}

func (cc *Client) SetRoomNo(roomNo int) {
	cc.roomNo = roomNo
}

func (cc *Client) GetUserName() string {
	return cc.userName
}

func (cc *Client) SetUserName(name string) {
	cc.userName = name
}

func (cc *Client) SetClientToAllow(client *Client) {
	cc.clientToAllow = client
}

func (cc *Client) GetClientToAllow() (result *Client) {
	return cc.clientToAllow
}

func (cc *Client) SetClientMsg(msg string) {
	cc.clientMsg = msg
}

func (cc *Client) GetClientMsg() (result string) {
	return cc.clientMsg
}

func (cc *Client) AddToList(client *Client, isAllowed bool) {
	if _, exists := cc.allowList[client]; !exists {
		cc.allowList[client] = isAllowed
	}
}

func (cc *Client) InTheList(client *Client) (result bool) {
	_, result = cc.allowList[client]

	return
}

func (cc *Client) IsAllowed(client *Client) (result bool) {
	result, _ = cc.allowList[client]

	return
}

func NewClient(connection net.Conn) *Client {
	writer := bufio.NewWriter(connection)
	reader := bufio.NewReader(connection)

	client := &Client{
		roomNo:    -1,
		msgin:     make(chan string),
		msgout:    make(chan string),
		msgAdmOut: make(chan string),
		allowList: make(map[*Client]bool),
		reader:    reader,
		writer:    writer,
	}

	client.Listen()

	return client
}
