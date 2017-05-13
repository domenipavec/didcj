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
	Name      string
	Ip        net.IP
	PrivateIp net.IP
	Username  string
	Password  string
}

type ServerByName []*Server

func (s ServerByName) Len() int           { return len(s) }
func (s ServerByName) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s ServerByName) Less(i, j int) bool { return s[i].Name < s[j].Name }
