package cores

import (
	"fmt"
	"net"
	"os"
	"time"
)

type Client struct {
	C    chan string //	用户发送数据的管道
	Name string      //	用户名
	Addr string      //	用户地址
}

// 保存在线用户	  cliAddr ===> client
var onlineMap map[string]Client

var message = make(chan string)

func PrintCp(str string) {
	fmt.Println("+-----------------------# chatroom #-----------------------+")
	_, err := os.Stdout.Write([]byte(str + "\n"))
	if err != nil {
		PrintCp(fmt.Sprint("os.Stdout.Write = ", err))
		return
	}
	fmt.Println("+------------------Welcome to PD CHATROOM------------------+")
	fmt.Println()
}

func MakeMsg(cli Client, msg string) (buf string) {
	curCliAddr := cli.Addr
	curCli := onlineMap[curCliAddr]
	pre := "+-----------------------# chatroom #-----------------------+\n"
	next := "+------------------Welcome to PD CHATROOM------------------+\n"
	buf = pre + "[" + curCli.Addr + "]" + curCli.Name + ": " + msg + "\n" + next
	return
}

func MakeSignalMsg(cli Client, msg string) (buf string) {
	curCliAddr := cli.Addr
	curCli := onlineMap[curCliAddr]
	buf = "[" + curCli.Addr + "]" + curCli.Name + ": " + msg + "\n"
	return
}

// HandleConn 处理用户连接
func HandleConn(conn net.Conn) {
	defer conn.Close()
	//	获取客户端的网路地址
	cliAddr := conn.RemoteAddr().String()
	//	创建一个结构体 Client, 默认用户名和网络地址一样
	cli := Client{make(chan string), cliAddr, cliAddr}
	//	把 cli 放入 map
	onlineMap[cliAddr] = cli
	//	新开一个协程，专门给当前客户端发送信息
	go WriteMagToClient(cli, conn)
	//	广播某个人在线
	message <- MakeMsg(cli, "login")
	//	提示我是谁
	cli.C <- MakeSignalMsg(cli, "I am here !")

	isQuit := make(chan bool)   //	对方是否主动退出
	isActive := make(chan bool) //	是否活跃

	//	新建一个协程，接收用户发送过来的数据
	go func() {
		buf := make([]byte, 2048)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				//	对方断开 或者 出问题
				PrintCp(fmt.Sprint("conn.Read err = ", err))
				isQuit <- true
				return
			}
			bufStr := string(buf[:n-1])
			msg := MakeSignalMsg(cli, bufStr)
			if len(bufStr) <= 0 {
				_, err := conn.Write([]byte("Don't send too fast !\n\n"))
				if err != nil {
					return
				}
				continue
			}
			//	过滤命令
			if bufStr[0] == '/' {
				Menu(conn, cli, bufStr)
			} else {
				//	转发此消息
				message <- msg
				//	报备服务器
			}
			isActive <- true //	表示活跃
		}
	}()

	for {
		//	通过 select检测 channel的流动
		select {
		case <-isQuit:
			message <- MakeMsg(cli, "login out") //	广播
			delete(onlineMap, cliAddr)           //	当前用户从 map移除
			return
		case <-isActive:

		case <-time.After(600 * time.Second): //	最大空闲活跃时间 	10 min
			delete(onlineMap, cliAddr)
			message <- MakeMsg(cli, "time out leave out")
			return
		}
	}
}

// WriteMagToClient	给当前客户端发送信息
func WriteMagToClient(cli Client, conn net.Conn) {
	PrintCp("[" + cli.Addr + "]" + cli.Name + ": login") //	报备服务器
	for msg := range cli.C {
		_, err := conn.Write([]byte(msg + "\n"))

		if err != nil {
			PrintCp(fmt.Sprint("给 ", cli.Addr, "发送消息失败!"))
		}
	}
}

// Manager	新开一个携程，来转发消息，只要有消息来了，就遍历 map，给 map的每一个成员都发送消息
func Manager() {
	//	给 map分配空间
	onlineMap = make(map[string]Client)
	for {
		msg := <-message //	没有消息前这里会阻塞
		//	遍历 map，给 map的每一个成员都发送消息
		for _, cli := range onlineMap {
			cli.C <- msg
		}
	}
}

func Core() {
	//	监听
	listener, err := net.Listen("tcp", ":7999")
	if err != nil {
		PrintCp(fmt.Sprint("net.Listen err = ", err))
		return
	}
	//	关闭监听
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			PrintCp("listener close fail!!")
		}
	}(listener)

	//	新开一个携程，来转发消息，只要有消息来了，就遍历 map，给 map的每一个成员都发送消息
	go Manager()

	//	主协程, 循环阻塞等待用户连接
	for {
		conn, err := listener.Accept()
		if err != nil {
			PrintCp(fmt.Sprint("listener.Accept err = ", err))
			continue
		}
		go HandleConn(conn) //	处理用户连接
	}

}
