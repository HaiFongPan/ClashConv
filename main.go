// Package main provides ...
package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

func readRemoteData(url string) string {
	r, err := http.Get(url)
	if err != nil {
		msg := "reading remote data error, " + err.Error()
		notification(msg)
		panic(msg)
	}

	defer r.Body.Close()
	body, _ := ioutil.ReadAll(r.Body)
	bodyString := prepareBase64(string(body))
	decode, _ := base64.StdEncoding.DecodeString(bodyString)
	return string(decode)
}

func decodeSs(content, include string, proxyRegs []*ProxyReg) ([]*Ss, []*ProxyGroup) {
	re := regexp.MustCompile(`//(.*)@(.*):(\d*)`)
	nameRe := regexp.MustCompile(`#(.*)`)
	obfsRe := regexp.MustCompile(`obfs=(.*);obfs-host=(.*)&`)

	nodes := strings.Split(content, "\n")
	r := []*Ss{}
	groups := []*ProxyGroup{}
	for _, v := range nodes {
		if len(strings.TrimSpace(v)) == 0 {
			continue
		}

		v, _ = url.PathUnescape(v)
		matchs := re.FindStringSubmatch(v)
		b64 := prepareBase64(matchs[1])
		cipherdecode, _ := base64.StdEncoding.DecodeString(b64)
		cipherstring := strings.Split(string(cipherdecode), ":")
		cipher := cipherstring[0]
		pswd := cipherstring[1]
		server := matchs[2]
		port := matchs[3]

		matchs = nameRe.FindStringSubmatch(v)
		namee := matchs[1]

		if include != "" && !filterNode(namee, include) {
			continue
		}

		n := &Ss{
			Type:     "ss",
			Server:   server,
			Port:     port,
			Cipher:   cipher,
			Password: pswd,
			Udp:      true,
		}
		n.Name = namee

		matchs = obfsRe.FindStringSubmatch(v)
		if len(matchs) > 2 {
			plugin := &SsPlugin{Mode: matchs[1], Host: matchs[2]}
			n.Plugin = "obfs"
			n.PluginOpt = plugin
		}
		r = append(r, n)
		appendToGroup(n.Name, proxyRegs)
	}

	for _, reg := range proxyRegs {
		if len(reg.group.Proxies) > 0 {
			groups = append(groups, reg.group)
		}
	}

	return r, groups
}

func appendToGroup(nodeName string, proxyRegs []*ProxyReg) {
	for _, reg := range proxyRegs {
		if len(reg.Pattern.FindStringSubmatch(nodeName)) > 0 {
			reg.group.Proxies = append(reg.group.Proxies, nodeName)
		}
	}
}

func decodeVmess(content, include string, proxyRegs []*ProxyReg) ([]*Vmess, []*ProxyGroup) {
	r := []*Vmess{}
	groups := []*ProxyGroup{}
	nodes := strings.Split(content, "\n")
	for _, v := range nodes {
		if len(strings.TrimSpace(v)) == 0 {
			continue
		}

		v = v[8:]
		b64 := prepareBase64(v)
		vmessJson, _ := base64.StdEncoding.DecodeString(b64)
		vmess := new(VmessC)
		json.Unmarshal(vmessJson, &vmess)

		if include != "" && !filterNode(vmess.Ps, include) {
			continue
		}
		alterId, _ := strconv.Atoi(vmess.Aid)
		n := &Vmess{
			Type:    "vmess",
			Server:  vmess.Add,
			Port:    vmess.Port,
			Uuid:    vmess.Id,
			AlterId: alterId,
			Cipher:  "auto",
			Udp:     true,
		}
		n.Name = vmess.Ps
		if vmess.Tls != "" {
			n.Tls = true
		}

		if vmess.Net == "ws" {
			n.Network = "ws"
			n.WSPath = vmess.Path
			header := make(map[string]string)
			header["Host"] = vmess.Host
			n.WSHeaders = header
		}

		r = append(r, n)
		appendToGroup(n.Name, proxyRegs)
	}

	for _, reg := range proxyRegs {
		if len(reg.group.Proxies) > 0 {
			groups = append(groups, reg.group)
		}
	}
	return r, groups
}

func prepareBase64(b string) string {
	e := (4 - len(b)%4) % 4
	for i := 0; i < e; i++ {
		b += "="
	}
	return b
}

func readTemplate(fileName string) string {
	c, err := ioutil.ReadFile(fileName)
	if err != nil {
		msg := "reading template error :" + err.Error()
		notification(msg)
		panic(msg)
	}
	return string(c)
}

func generate(cnf *Cnf) ([]string, []string) {
	nodeConfs := []string{}
	groupConfs := []string{}
	nodeNames := []string{}
	groups := []*ProxyGroup{}

	for _, v := range cnf.Airports {
		active := false
		for _, a := range cnf.Active {
			if a == v.Name {
				active = true
			}
		}

		if !active {
			continue
		}

		data := readRemoteData(v.Subs)
		if v.proto == SS {
			ns, gs := decodeSs(data, v.IncludeReg, v.regs)
			groups = append(groups, gs...)
			for _, n := range ns {
				j, _ := json.Marshal(n)
				nodeConfs = append(nodeConfs, fmt.Sprintf(" - %s\n", string(j)))
				nodeNames = append(nodeNames, n.Name)
			}
		} else if v.proto == VMESS {
			ns, gs := decodeVmess(data, v.IncludeReg, v.regs)
			groups = append(groups, gs...)
			for _, n := range ns {
				j, _ := json.Marshal(n)
				nodeConfs = append(nodeConfs, fmt.Sprintf(" - %s\n", string(j)))
				nodeNames = append(nodeNames, n.Name)
			}
		}
	}

	groups = makeGroups(nodeNames, groups)

	for _, n := range groups {
		j, _ := json.Marshal(n)
		groupConfs = append(groupConfs, fmt.Sprintf(" - %s\n", string(j)))
	}
	return nodeConfs, groupConfs
}

func filterNode(name string, includePattern string) bool {
	re := regexp.MustCompile(includePattern)
	find := re.FindStringSubmatch(name)
	return len(find) > 0
}

func formatConf(cnf *Cnf, nodeConfs []string, groupConfs []string) {
	var nbuf bytes.Buffer
	var gbuf bytes.Buffer

	for _, v := range nodeConfs {
		nbuf.WriteString(v)
	}

	for _, v := range groupConfs {
		gbuf.WriteString(v)
	}

	template := readTemplate(cnf.TemplatePath)
	config := strings.Replace(template, "[proxy_config]", nbuf.String(), 1)
	config = strings.Replace(config, "[proxy_group_configs]", gbuf.String(), 1)
	ioutil.WriteFile(cnf.TargetPath, []byte(config), 0666)
	notification("更新成功")
}

func notification(message string) {
	cmd := fmt.Sprintf("display notification \"%s\" with title \"Clash 配置更新\"", message)
	exec.Command("/usr/bin/osascript", "-e", cmd).Run()
}

var (
	configJson string
)

func main() {
	flag.StringVar(&configJson, "f", "", "set config file")
	flag.Parse()
	cnf := ReadConf(configJson)
	nodes, groups := generate(cnf)
	formatConf(cnf, nodes, groups)
}
