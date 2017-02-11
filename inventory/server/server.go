package server

import "net"

type Server struct {
	Ip       net.IP
	Username string
	Password string
}
