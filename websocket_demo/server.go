package websocket_demo

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type msg struct {
	timeAfter int64
	data      string
}

func parse(b []byte, m *msg) error {
	str := string(b)
	ss := strings.Split(strings.TrimSpace(str), ",")
	if len(ss) != 2 {
		return errors.New("invalid input:" + str)
	}
	t, err := strconv.ParseInt(ss[0], 10, 32)
	if err != nil {
		return err
	}
	m.timeAfter = t
	m.data = ss[1]
	return nil
}

func Start() {
	address := "localhost:8081"

	// 创建一个新的 http.Server
	server := &http.Server{
		Addr:    address,
		Handler: nil, // 使用默认的多路复用器
	}

	// 设置路由处理函数
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		fmt.Println("1111")
		http.ServeFile(writer, request, "websocket_demo/home.html")
	})
	http.HandleFunc("/ws", func(writer http.ResponseWriter, request *http.Request) {
		serverWs(writer, request)
	})

	// 启动服务器
	go func() {
		fmt.Printf("Starting server at %s\n", address)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("ListenAndServe error: %v\n", err)
		}
	}()

	// 创建一个通道来监听系统中断信号
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	// 阻塞等待中断信号
	<-stop

	// 创建一个 5 秒的上下文用于关闭服务器
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fmt.Println("Shutting down server...")

	// 优雅地关闭服务器
	if err := server.Shutdown(ctx); err != nil {
		fmt.Printf("Server Shutdown Failed:%+v", err)
	}
	fmt.Println("Server exited")
}

func serverWs(writer http.ResponseWriter, request *http.Request) error {
	conn, err := upgrader.Upgrade(writer, request, nil)
	if err != nil {
		return err
	}
	go read(conn)
	return nil
}

func read(conn *websocket.Conn) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			break
		}
		data := &msg{}
		if err := parse(message, data); err != nil {
			continue
		}
		timer := time.NewTimer(time.Duration(data.timeAfter) * time.Second)
		go func(data *msg, conn *websocket.Conn) {
			<-timer.C
			// 在定时器到期后执行的操作
			w, err := conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write([]byte(data.data))
			fmt.Printf("Message content: %s\n", data.data)
			if err := w.Close(); err != nil {
				return
			}
		}(data, conn)
	}
}

func writer(coon *websocket.Conn) {

}
