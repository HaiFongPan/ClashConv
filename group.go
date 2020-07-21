package main

import "regexp"

type ProxyGroup struct {
	Name     string   `json:"name,omitempty"`
	Type     string   `json:"type,omitempty"`
	Proxies  []string `json:"proxies,omitempty"`
	Url      string   `json:"url,omitempty"`
	Interval int      `json:"interval,omitempty"`
}

type ProxyReg struct {
	Pattern *regexp.Regexp
	group   *ProxyGroup
}

func newUrlTestGroup(name string) *ProxyGroup {
	g := &ProxyGroup{
		Type:     "url-test",
		Name:     name,
		Proxies:  []string{},
		Url:      "http://www.gstatic.com/generate_204",
		Interval: 3600,
	}
	return g
}

func makeGroups(nodes []string, testUrls []*ProxyGroup) []*ProxyGroup {
	groups := []*ProxyGroup{}
	// proxy
	proxyRule := &ProxyGroup{
		Type:    "select",
		Name:    "ProxyRule",
		Proxies: []string{"DIRECT"},
	}
	for _, v := range testUrls {
		proxyRule.Proxies = append(proxyRule.Proxies, v.Name)
	}
	for _, v := range nodes {
		proxyRule.Proxies = append(proxyRule.Proxies, v)
	}
	// spotify
	groups = append(groups, proxyRule)
	groups = append(groups, newSelectGroup("Blacklist", testUrls))
	groups = append(groups, newSelectGroup("AppleRule", testUrls))
	groups = append(groups, newSelectGroup("MediaRule", testUrls))
	groups = append(groups, newSelectGroup("DlerRule", testUrls))
	groups = append(groups, newSelectGroup("Spotify", testUrls))
	groups = append(groups, newSelectGroup("Hijack", testUrls))
	groups = append(groups, testUrls...)

	return groups
}

func newSelectGroup(name string, testUrls []*ProxyGroup) *ProxyGroup {
	g := &ProxyGroup{
		Type:    "select",
		Name:    name,
		Proxies: []string{"DIRECT", "REJECT", "ProxyRule"},
	}
	for _, v := range testUrls {
		g.Proxies = append(g.Proxies, v.Name)
	}
	return g
}
