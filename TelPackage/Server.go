package TelPackage

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip        string
	Port      int
	OnlineMap map[string]*TestUser
	Maplock   sync.RWMutex
	Message   chan string
}

func NewServer(ip string, port int) *Server {
	Server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*TestUser),
		Message:   make(chan string),
	}
	return Server
}

func (srv *Server) ListenMessage() {
	for {
		msg := <-srv.Message

		srv.Maplock.Lock()
		for _, cli := range srv.OnlineMap {
			cli.UserChan <- msg
		}
		srv.Maplock.Unlock()
	}
}

func (srv *Server) BroadCast(user *TestUser, msg string) {
	sendMSg := "[" + user.Addr + "]" + user.Name + ":" + msg
	srv.Message <- sendMSg
}

func (srv *Server) Handler(conn net.Conn) {
	teluser := NewUser(conn, srv)
	teluser.Online()

	isonline := make(chan bool)

	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				teluser.Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err:", err)
				return
			}

			msg := string(buf[:n-1])

			teluser.DoMessage(msg)

			isonline <- true
		}
	}()
	fmt.Println("Now UserNumber: ", len(srv.OnlineMap))
	for {
		select {
		case <-isonline:

		case <-time.After((time.Second * 150)):
			teluser.SendMsg("You were forced off the line")

			close(teluser.UserChan)

			conn.Close()

			return
		}
	}
}

func (srv *Server) Start() {
	Listener, err := net.Listen("tcp4", fmt.Sprintf("%s:%d", srv.Ip, srv.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}
	defer Listener.Close()

	go srv.ListenMessage()

	for {
		conn, err := Listener.Accept()
		if err != nil {
			fmt.Println("listener accept err:", err)
			continue
		}
		go srv.Handler(conn)
	}
}
