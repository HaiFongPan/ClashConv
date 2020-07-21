package main

type Proto int

const (
	SS Proto = iota
	VMESS
)

type Node struct {
	Name string `json:"name"`
}

type VmessC struct {
	Add  string `json:"add"`
	Aid  string `json:"aid"`
	Id   string `json:"id"`
	Net  string `json:"net"`
	Path string `json:"path"`
	Port int    `json:"port"`
	Ps   string `json:"ps"`
	Tls  string `json:"tls"`
	Type string `json:"type"`
	V    string `json:"v"`
	Host string `json:"host"`
}

type Vmess struct {
	Node
	Type      string            `json:"type"`
	Server    string            `json:"server"`
	Port      int               `json:"port"`
	Uuid      string            `json:"uuid"`
	AlterId   int               `json:"alterId"`
	Cipher    string            `json:"cipher"`
	Udp       bool              `json:"udp,omitempty"`
	Tls       bool              `json:"tls,omitempty"`
	Network   string            `json:"network,omitempty"`
	WSPath    string            `json:"ws-path,omitempty"`
	WSHeaders map[string]string `json:"ws-headers,omitempty"`
}

type Ss struct {
	Node
	Type      string    `json:"type"`
	Server    string    `json:"server"`
	Port      string    `json:"port"`
	Cipher    string    `json:"cipher"`
	Password  string    `json:"password"`
	Udp       bool      `json:"udp"`
	Plugin    string    `json:"plugin,omitempty"`
	PluginOpt *SsPlugin `json:"plugin-opts,omitempty"`
}

type SsPlugin struct {
	Mode string `json:"mode"`
	Host string `json:"host"`
}
