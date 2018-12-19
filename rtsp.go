// 用于处理RTSP的请求协议
package main

import (
	"fmt"
	"net"
)

// RTSPRequest 用于解析RTSP请求头的参数
type RTSPRequest struct {
	method string
	url    string
}

func main() {
	listen, err := net.Listen("tcp", ":554")
	if err != nil {
		fmt.Println("listen error: ", err)
		return
	}

	for {
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println("accept error:", err)
			break
		}

		go HandleConn(conn)
	}
}

// HandleConn 提供对RTSP协议的解析
func HandleConn(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 1024)

	for {
		fmt.Println("start to read from conn ...")
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("conn read error: ", err)
			continue
		}
		fmt.Printf("read %d bytes, content is %s\n", n, string(buf[:n]))
		var rtspReq RTSPRequest
		// 解析RTSP的方法
		{
			rtspReq.method = "OPTIONS"
		}
		switch rtspReq.method {
		case "OPTIONS":
			{
				HandleOptions(conn)
			}
		default:
			{
				fmt.Println("Unkown Method!")
			}
		}
	}
}

// HandleOptions 处理RTSP中的OPTIONS方法
func HandleOptions(conn net.Conn) {
	fmt.Println("Handle OPTIONS!")
	conn.Write([]byte("options methods"))
}
