package main

//对外接口

const ()

type IConn interface {
	Send(msgType int, msg []byte) //发送消息
	Close()                       //关闭连接
}
