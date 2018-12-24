// Package rtspgo 是一个RTSP处理模块
package rtspgo

import (
	"fmt"
	"net"
	"strings"
)

// contextKey 是用于context.WithValue的使用
type contextKey struct {
	name string
}

func (c *contextKey) String() string {
	return "rtsp context value " + c.name
}

// RTSPRequest 用于解析RTSP请求头的参数
type RTSPRequest struct {
	method string
	url    string
	cseq   string
}

// RtspMethodType RTSP支持的所有方法
var RtspMethodType = []string{"OPTIONS", "DESCRIBE", "SETUP", "PLAY", "PAUSE", "TEARDOWN"}

// HandleConn 提供对RTSP协议的解析
func HandleConn(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 1024)

	for {
		fmt.Println("start to read from conn ...")
		_, err := conn.Read(buf)
		if err != nil {
			fmt.Println("conn read error: ", err)
			continue
		}
		//fmt.Printf("read %d bytes, content is %s\n", n, string(buf[:n]))
		var rtspReq RTSPRequest
		// 解析RTSP的方法
		{
			strReq := string(buf)
			sliceStrReq := strings.Split(string(strReq), "\r\n")
			// 对首行方法名进行解析
			sliceFirstLine := strings.Split(string(sliceStrReq[0]), " ")
			rtspReq.method = sliceFirstLine[0]
			i, err := FindStrFromSlice(rtspReq.method, RtspMethodType)
			if err != nil || i == -1 {
				fmt.Println("unkown Method")
				continue
			}
			// 对CSeq进行解析
		}
		switch rtspReq.method {
		case "OPTIONS":
			{
				HandleOptions(conn, rtspReq)
			}
		default:
			{
				fmt.Println("Unkown Method!")
			}
		}
	}
}

// HandleOptions 处理RTSP中的OPTIONS方法
func HandleOptions(conn net.Conn, info RTSPRequest) {
	fmt.Println("Handle OPTIONS!")
	conn.Write([]byte("options methods"))
}
