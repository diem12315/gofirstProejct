package TelPackage

import (
	"net"
)

type TestUser struct {
	Name     string
	Addr     string
	UserChan chan string
	conn     net.Conn
	server   *Server
}

func NewUser(conn net.Conn, server *Server) *TestUser {
	userAddr := conn.RemoteAddr().String()

	user := &TestUser{
		Name:     userAddr,
		Addr:     userAddr,
		UserChan: make(chan string),
		conn:     conn,
		server:   server,
	}

	go user.ListenMessage()

	return user
}

func (usr *TestUser) Online() {
	usr.server.Maplock.Lock()
	usr.server.OnlineMap[usr.Name] = usr
	usr.server.Maplock.Unlock()
	usr.server.BroadCast(usr, "online")

}

func (usr *TestUser) Offline() {
	usr.server.Maplock.Lock()
	delete(usr.server.OnlineMap, usr.Name)
	usr.server.Maplock.Unlock()

	usr.server.BroadCast(usr, "offfline")
}

func (usr *TestUser) DoMessage(msg string) {
	usr.server.BroadCast(usr, msg)
}

func (usr *TestUser) ListenMessage() {
	for {
		msg := <-usr.UserChan
		usr.conn.Write([]byte(msg + "\n"))
	}
}
