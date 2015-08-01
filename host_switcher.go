// Copyright 2014 Unknwon
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

// Package switcher is a helper module that provides host switch functionality for Macaron.
package switcher

import (
	"log"
	"net/http"
	"strings"

	"github.com/Unknwon/com"
	"github.com/Unknwon/macaron"
)

type matchType int

const (
	_PLAIN matchType = iota
	_SUFFIX
)

type host struct {
	m    *macaron.Macaron
	name string
	matchType
}

// HostSwitcher represents a global multi-site support layer.
type HostSwitcher struct {
	hosts []*host
}

// NewHostSwitcher initalizes and returns a new host switcher.
// You have to use this function to get a new host switcher.
func NewHostSwitcher() *HostSwitcher {
	return &HostSwitcher{}
}

// Set adds a new switch to host switcher.
func (hs *HostSwitcher) Set(name string, m *macaron.Macaron) {
	if len(name) == 0 {
		return
	}

	h := &host{m, name, _PLAIN}
	if name[0] == '*' {
		h.name = h.name[1:]
		h.matchType = _SUFFIX
	}
	hs.hosts = append(hs.hosts, h)
}

// Remove removes a switch from host switcher.
func (hs *HostSwitcher) Remove(name string) {
	idx := -1
	for i, h := range hs.hosts {
		if h.name == name {
			idx = i
			break
		}
	}
	if idx >= 0 {
		hs.hosts = append(hs.hosts[:idx], hs.hosts[idx+1:]...)
	}
}

// ServeHTTP is the HTTP Entry point for a Host Switcher instance.
func (hs *HostSwitcher) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	for i := range hs.hosts {
		switch hs.hosts[i].matchType {
		case _PLAIN:
			if hs.hosts[i].name == req.Host {
				hs.hosts[i].m.ServeHTTP(resp, req)
				return
			}
		case _SUFFIX:
			if strings.HasSuffix(req.Host, hs.hosts[i].name) {
				hs.hosts[i].m.ServeHTTP(resp, req)
				return
			}
		}
	}

	http.Error(resp, "Not Found", http.StatusNotFound)
}

// RunOnAddr runs server in given address and port.
func (hs *HostSwitcher) RunOnAddr(addr string) {
	if macaron.Env == macaron.DEV {
		infos := strings.Split(addr, ":")
		port := com.StrTo(infos[1]).MustInt()
		for i := range hs.hosts[:len(hs.hosts)-1] {
			go hs.hosts[i].m.Run(infos[0], port)
			port++
		}
		hs.hosts[len(hs.hosts)-1].m.Run(infos[0], port)
		return
	}

	log.Fatalln(http.ListenAndServe(addr, hs))
}

// GetDefaultListenAddr returns default server listen address of Macaron.
func GetDefaultListenAddr() string {
	host, port := macaron.GetDefaultListenInfo()
	return host + ":" + com.ToStr(port)
}

// Run the http server. Listening on os.GetEnv("PORT") or 4000 by default.
func (hs *HostSwitcher) Run() {
	hs.RunOnAddr(GetDefaultListenAddr())
}
