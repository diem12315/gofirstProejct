package TelPackage

import (
	"net"
	"strings"
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

func (usr *TestUser) SendMsg(msg string) {
	usr.conn.Write([]byte(msg))
}

func (usr *TestUser) DoMessage(msg string) {
	if msg == "who" {
		usr.server.Maplock.Lock()
		for _, user := range usr.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + ":" + "online...\n"
			usr.SendMsg(onlineMsg)
		}
		usr.server.Maplock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		newName := strings.Split(msg, "|")[1]
		_, ok := usr.server.OnlineMap[newName]
		if ok {
			usr.SendMsg("当前用户名被占用\n")
		} else {
			usr.server.Maplock.Lock()
			delete(usr.server.OnlineMap, usr.Name)
			usr.server.OnlineMap[newName] = usr
			usr.server.Maplock.Unlock()
			usr.Name = newName
			usr.SendMsg("你已经更新用户名" + usr.Name + "\n")
		}
	} else if len(msg) > 4 && msg[:3] == "to|" {
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			usr.SendMsg("消息格式不正确，请使用 \"to|张三|你好啊\"格式。 \n")
			return
		}
		remoteUser, ok := usr.server.OnlineMap[remoteName]
		if !ok {
			usr.SendMsg("该用户不存在\n")
			return
		}

		content := strings.Split(msg, "|")[2]
		if content == " " {
			usr.SendMsg("无消息内容，请重发\n")
			return
		}
		remoteUser.SendMsg(usr.Name + "对你说：" + content + "\n")

	} else {
		usr.server.BroadCast(usr, msg)
	}
}

func (usr *TestUser) ListenMessage() {
	for {
		msg := <-usr.UserChan
		usr.conn.Write([]byte(msg + "\n"))
	}
}
