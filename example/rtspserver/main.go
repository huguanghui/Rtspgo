package main

import (
	"fmt"
	"net"

	rtspgo "github.com/huguanghui/Rtspgo"
)

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

		go rtspgo.HandleConn(conn)
	}
}
