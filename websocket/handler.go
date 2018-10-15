package ws

import (
	"context"
	"html/template"
	"log"
	"net/http"
)

//对外的API
//获取客户端页面
func ClientHandler(w http.ResponseWriter, r *http.Request) {
	s, err := Asset("websocket/client.html")
	if err != nil {
		return
	}

	tmpl := template.New("client")
	tmpl.Parse(string(s))
	tmpl.Execute(w, "ws client")
}

//单独推送消息
func PushHandler(w http.ResponseWriter, r *http.Request) {
	//取出推送给的连接
	wscon := GetConnByFd(2)
	if wscon == nil {
		w.Write([]byte("not exist"))
		return
	}
	wscon.Send(1, []byte("my name is meng"))
	w.Write([]byte("success"))
	return
}

//广播消息
func BroadcastHandler(w http.ResponseWriter, r *http.Request) {
	for _, wscon := range GetAllConn() {
		if wscon != nil {
			wscon.Send(1, []byte("broadcast"))
		}
	}

	w.Write([]byte("success"))
	return

}

func WSHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(context.Background()) //取消

	con, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("upgrade error:%v", err)
		return
	}
	//初始化一个ws连接结构
	wsConn := initWsConn(con)

	//读取消息
	go wsConn.read(ctx)
	//写入消息
	go wsConn.write(ctx)
	//关闭
	go wsConn.close(cancel)

}
