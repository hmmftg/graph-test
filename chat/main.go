package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

var allClients map[string]*Client

type Chat struct {
	Message string
	Sender  *Client
	To      string
}

func (c Chat) String() string {
	return fmt.Sprintf("\nFrom: %s, Text: %s\n%s", c.Sender.ID, c.Message, Prompt)
}

const (
	From   = "From: "
	To     = "To:"
	Prompt = "Enter your target:"
)

type Client struct {
	ID          string
	incoming    chan *Chat
	reader      *bufio.Reader
	writer      *bufio.Writer
	conn        net.Conn
	connections map[string]*Client
}

func (client *Client) Read() {
	for {
		err := write(client.writer, Prompt)
		if err != nil {
			log.Println("unable to write", err)
			break
		}
		target, err := client.reader.ReadString('\n')
		if err != nil {
			log.Println(client.ID, client.Name(), err)
			break
		}
		target = target[:len(target)-2]
		to, available := allClients[target]
		if !available {
			log.Println(target, "not available")
			continue
		}
		err = write(client.writer, "Enter your chat:")
		if err != nil {
			log.Println("unable to write", err)
			break
		}
		text, err := client.reader.ReadString('\n')
		if err != nil {
			log.Println(client.ID, client.Name(), err)
			break
		}
		ch := Chat{
			To:      target,
			Message: text[:len(text)-2],
			Sender:  client,
		}
		log.Println(client.ID, "request sending", ch.Message, "to", to.ID)
		to.incoming <- &ch
	}

	client.conn.Close()
	delete(allClients, client.Name())
	client = nil
}

func (client *Client) Name() string {
	return client.conn.RemoteAddr().String()
}

func write(target *bufio.Writer, text string) error {
	_, err := target.WriteString(text)
	if err != nil {
		return err
	}
	target.Flush()
	return nil

}

func (client *Client) Write() {
	for {
		chat := <-client.incoming
		log.Println(chat.Sender.ID, "sending", chat.Message, "to", client.ID)
		err := write(client.writer, chat.String())
		if err != nil {
			log.Println(client.ID, client.Name(), "unable to write chat", err)
			return
		}
	}
}

func NewClient(connection net.Conn) *Client {
	writer := bufio.NewWriter(connection)
	reader := bufio.NewReader(connection)

	err := write(writer, "Enter your id:")
	if err != nil {
		log.Println("unable to write", err)
		return nil
	}
	writer.Flush()

	id, err := reader.ReadString('\n')
	if err != nil {
		log.Println("unable to get client id", err)
		return nil
	}

	client := &Client{
		ID:          id[:len(id)-2],
		incoming:    make(chan *Chat),
		conn:        connection,
		reader:      reader,
		writer:      writer,
		connections: make(map[string]*Client),
	}
	go client.Read()
	go client.Write()

	allClients[client.ID] = client
	log.Println(connection.RemoteAddr().String(), "connected as", client.ID)

	return client
}

func main() {
	allClients = make(map[string]*Client)
	listener, _ := net.Listen("tcp", ":8080")
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err.Error())
		}
		go NewClient(conn)
	}
}
