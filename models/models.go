package models

import "net"

type Report struct {
	Ip         string   `json:"ip"`
	Messages   []string `json:"messages"`
	SendCount  int      `json:"send_count"`
	LargestMsg int      `json:"largest_msg"`
	RunTime    int64    `json:"run_time"`
	MaxMemory  int      `json:"max_memory"`
}

type Server struct {
	Ip        net.IP
	PrivateIp net.IP
	Username  string
	Password  string
}
