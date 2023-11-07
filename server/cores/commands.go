package cores

import (
	"fmt"
	"net"
	"regexp"
	"strings"
)

func Menu(conn net.Conn, cli Client, command string) {
	switch {
	case command == "/who":
		Who_(conn, cli)
	case strings.Contains(command, "/rename"):
		//	报备服务器 ， 谁 使用了 什么命令
		PrintCp(MakeSignalMsg(cli, command))
		if len(command) < 9 || command[:8] != "/rename|" {
			_, err := conn.Write([]byte("@ Format error, should be \"/rename|[yourname]\"\n\n"))
			if err != nil {
				PrintCp(fmt.Sprint("conn.Write err = ", err))
			}
			return
		}
		//	判断名字是否合法
		name := strings.Split(command, "|")[1]
		if NameIsFormat(name) == false {
			_, err := conn.Write([]byte("@ Your name is limited to: numbers, letters, underscores 4 ~ 16\n\n"))
			if err != nil {
				PrintCp(fmt.Sprint("conn.Write err = ", err))
			}
			return
		}
		Rename_(conn, cli, command)
	default:
		ERROR(conn, cli, command)
	}
}

func NameIsFormat(name string) bool {
	re, err := regexp.Compile("[A-Za-z0-9_]{4,16}")
	if err != nil {
		PrintCp(fmt.Sprint("NameIsFormat err = ", err))
		return false
	}
	return re.MatchString(name)
}

func ERROR(conn net.Conn, cli Client, command string) {
	PrintCp(MakeSignalMsg(cli, "error_command = "+command))
	msg := "unknown command !\n\n"
	_, err := conn.Write([]byte(msg))
	if err != nil {
		PrintCp(fmt.Sprint("conn.Write err = ", err))
		return
	}
}

func Who_(conn net.Conn, cli Client) {
	//	报备服务器 ， 谁 使用了 什么命令
	PrintCp(MakeSignalMsg(cli, "/who"))
	//	遍历 map
	_, err := conn.Write([]byte("user list:\n"))
	if err != nil {
		PrintCp(fmt.Sprint("conn.Write err = ", err))
		return
	}
	for _, tmp := range onlineMap {
		msg := tmp.Addr + ":" + tmp.Name + "\n"
		_, err := conn.Write([]byte(msg))
		if err != nil {
			PrintCp(fmt.Sprint("conn.Write err = ", err))
			return
		}
	}
	_, err = conn.Write([]byte("\n"))
	if err != nil {
		PrintCp(fmt.Sprint("conn.Write err = ", err))
		return
	}
}

func Rename_(conn net.Conn, cli Client, command string) {
	name := strings.Split(command, "|")[1]
	cli.Name = name
	onlineMap[cli.Addr] = cli
	_, err := conn.Write([]byte("@ rename success!\n\n"))
	if err != nil {
		PrintCp(fmt.Sprint("conn.Write err = ", err))
		return
	}
}
