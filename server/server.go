package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

const (
	ReqCmdCode = "cmd"

	Telnet = 2 //For testing
)

type Server struct {
	*Router
	addr string
}

func ListenAndServe(addr string, router *Router) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	defer listener.Close()

	s := Server{
		addr:   addr,
		Router: router,
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		go s.HandleReq(conn)
	}
}

func (s *Server) HandleReq(conn net.Conn) {
	buff := make([]byte, 4096)
	defer conn.Close()

	n, err := conn.Read(buff)
	if err != nil {
		log.Println(err)
		return
	}

	ctx := &Ctx{
		Response: Response{Conn: conn},
		header:   ParserReq(buff[:n-Telnet]),
	}
	cmd := ctx.header[ReqCmdCode]

	s.RLock()
	defer s.RUnlock()

	if fn, ok := s.Router.fn[cmd]; ok {
		fn(ctx)
	} else {
		conn.Write([]byte("\r wrong command\n"))
	}
}

func ParserReq(payload []byte) map[string]string {
	header := make(map[string]string)

	scanner := bufio.NewScanner(strings.NewReader(string(payload)))
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println(line)

		if strings.TrimSpace(line) == "" {
			continue
		}

		parts := strings.Split(line, ":")
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			header[key] = value
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Parsing error:", err)
	}

	return header
}
