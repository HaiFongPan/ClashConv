package main

import (
	"encoding/json"
	"io/ioutil"
	"regexp"
)

type Cnf struct {
	TargetPath   string     `json:"targetPath"`
	TemplatePath string     `json:"templatePath"`
	Airports     []*Airport `json:"airports"`
	Active       []string   `json:"active"`
}

type Airport struct {
	Subs       string      `json:"subs"`
	Proto      string      `json:"proto"`
	Name       string      `json:"name"`
	IncludeReg string      `json:"includeReg"`
	Groups     []*GroupCnf `json:"groups"`
	proto      Proto
	regs       []*ProxyReg
}

type GroupCnf struct {
	Reg       string `json:"reg"`
	GroupName string `json:"groupName"`
}

func ReadConf(path string) *Cnf {
	bs, _ := ioutil.ReadFile(path)
	cnf := &Cnf{}
	if err := json.Unmarshal(bs, cnf); err != nil {
		panic("unmarshal config error :" + err.Error())
	}

	for _, conf := range cnf.Airports {
		if conf.Proto == "shadowsocks" {
			conf.proto = SS
		} else {
			conf.proto = VMESS
		}

		for _, rg := range conf.Groups {
			conf.regs = append(conf.regs, &ProxyReg{regexp.MustCompile(rg.Reg), newUrlTestGroup(rg.GroupName)})
		}
	}
	return cnf
}
