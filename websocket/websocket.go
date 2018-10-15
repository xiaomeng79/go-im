package ws

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
)

//type Fd int64

//存储全局的连接
var (
	fd2conn sync.Map
	conn2fd sync.Map
	fd      int64 //递增
)

//TODO:将协程异常做处理

//API推送消息
func (ws *wsConn) Send(msgType int, msg []byte) {
	//TODO:如果缓存消息满了，丢弃或者怎么处理
	ws.outChan <- &wsMessage{
		messageType: msgType,
		data:        msg,
	}
}

//主动关闭
func (ws *wsConn) Close() {
	ws.closeChan <- struct{}{}
}

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

//客户端读写消息
type wsMessage struct {
	messageType int
	data        []byte
}

//websocket连接
type wsConn struct {
	conn      *websocket.Conn //底层websocket连接
	inChan    chan *wsMessage //读消息的缓存，客户端发送来的消息
	outChan   chan *wsMessage //写消息的缓存，服务端发送的消息
	closeChan chan struct{}   //关闭通道
}

//websocket实现连接

func initWsConn(con *websocket.Conn) *wsConn {
	wc := &wsConn{
		conn:      con,
		inChan:    make(chan *wsMessage, 1024),
		outChan:   make(chan *wsMessage, 1024),
		closeChan: make(chan struct{}, 1),
	}
	//设置关闭函数
	wc.conn.SetCloseHandler(func(code int, text string) error {
		fmt.Println("client close")
		wc.closeChan <-struct {}{}
		return nil
	})
	wc.add() //增加对应关系
	return wc

}

//读取连接信息，写入缓存
func (ws *wsConn) read(ctx context.Context) {
	for {
		mtp, mdata, err := ws.conn.ReadMessage()
		if err != nil {
			fmt.Println(err.Error())
			ws.conn.WriteMessage(websocket.CloseMessage,[]byte("read error")) //关闭连接
			ws.closeChan <- struct{}{}
			fmt.Println("read message error")
			return
		}
		//TODO:判断消息是否阻塞，丢弃或者做其他处理

		//将消息放入inchan
		ws.inChan <- &wsMessage{
			messageType: mtp,
			data:        mdata,
		}
		//TODO:测试发送回去,使用的时候删除
		ws.outChan <- &wsMessage{
			messageType: mtp,
			data:        mdata,
		}
		//关闭
		select {
		case <-ctx.Done():
			log.Println("read close")
			return
		default:

		}
	}
}

//读取通道缓冲消息，写入连接
func (ws *wsConn) write(ctx context.Context) {
	for {
		select {
		case msg := <-ws.outChan:
			err := ws.conn.WriteMessage(msg.messageType, msg.data)
			if err != nil {
				ws.conn.WriteMessage(websocket.CloseMessage,[]byte("write error")) //关闭连接
				ws.closeChan <- struct{}{}
				fmt.Println("write message error")
				return
			}
		case <-ctx.Done():
			log.Println("write close")
			return
		}
	}
}

//关闭
func (ws *wsConn) close(cancel context.CancelFunc) {
	defer ws.conn.Close()
	<-ws.closeChan
	ws.del() //删除对应关系
	cancel()
}

/************************************fd*********************************/

//增加
func (ws *wsConn) add() {
	atomic.AddInt64(&fd, 1)
	fmt.Println("fd:", fd)
	fd2conn.Store(fd, ws)
	conn2fd.Store(ws, fd)
}

//减少
func (ws *wsConn) del() {
	if v, ok := conn2fd.Load(ws); ok {
		fd2conn.Delete(v)
		conn2fd.Delete(ws)
	}
}

//根据fd获取conn
func GetConnByFd(fd int64) *wsConn {
	v, ok := fd2conn.Load(fd)
	if !ok {
		log.Println("fd not exist")
		return nil
	}
	return v.(*wsConn)
}

//获取全部
func GetAllConn() []*wsConn {
	wss := make([]*wsConn, 1)
	conn2fd.Range(func(key, value interface{}) bool {
		wss = append(wss, key.(*wsConn))
		return true
	})
	return wss

}
