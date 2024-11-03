package server

import "net"

type Response struct {
	net.Conn
}

func (r *Response) ResWriter(msg string) error {
	_, err := r.Write([]byte(msg))
	return err
}
