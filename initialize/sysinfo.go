package main

import (
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"unicode"

	"github.com/fourth04/initialize/utils"
	ps "github.com/mitchellh/go-ps"
)

type Adapter struct {
	name       string
	flags      []string
	mtu        int
	inet       string
	netmask    string
	broadcast  string
	inet6      string
	prefixlen  int
	ether      string
	txqueuelen int
}

var CRLF = getCRLF()

func getCRLF() string {
	switch runtime.GOOS {
	case "windows":
		return "\r\n"
	case "linux":
		return "\n"
	default:
		return "\n"
	}
}

func getIfInfo() []*Adapter {
	// bytes, _ := utils.ReadFileFast("..\\docs\\ifconfig.txt")
	// stdout := string(bytes)
	// stdout, _, _ := utils.ExecuteAndGetResult("cat ..\\docs\\ifconfig.txt")
	stdout, _, _ := utils.ExecuteAndGetResult("ifconfig -a")

	var pat string

	var rv []*Adapter
	pat = "(?m)" + CRLF + CRLF
	reg, _ := regexp.Compile(pat)
	splitedSlice := reg.Split(stdout, -1)
	for _, ifInfo := range splitedSlice {
		if strings.TrimSpace(ifInfo) == "" {
			continue
		}
		adapter := Adapter{}
		ifInfoSplited := strings.Split(ifInfo, CRLF)

		for _, line := range ifInfoSplited {
			var words []string
			for _, word := range strings.Split(line, " ") {
				if word != "" {
					words = append(words, word)
				}
			}
			switch {
			case strings.Contains(line, "mtu"):
				adapter.name = strings.Trim(words[0], ":")
				adapter.flags = strings.Split(strings.Trim(strings.Split(words[1], "<")[1], ">"), ",")
				adapter.mtu, _ = strconv.Atoi(words[len(words)-1])
			case strings.Contains(line, "inet6"):
				adapter.inet6 = words[1]
				adapter.prefixlen, _ = strconv.Atoi(words[3])
			case strings.Contains(line, "inet"):
				adapter.inet = words[1]
				adapter.netmask = words[3]
				if len(words) == 6 {
					adapter.broadcast = words[5]
				}
			case strings.Contains(line, "ether"):
				adapter.ether = words[1]
				adapter.txqueuelen, _ = strconv.Atoi(words[3])
			}
		}
		rv = append(rv, &adapter)
	}
	// log.Println(len(reg.Split(string(bytes), -1)))
	return rv
}

func getDpdkDevBind() []string {
	stdout, stderr, err := utils.ExecuteAndGetResult("paramsdpdk-devbind.py --status | grep 'drv=igb_uio' | awk '{print $1}'")
	if stderr != "" || err != nil {
		return []string{}
	}
	return strings.Split(stdout, "\n")
}

func getProcesses(processName string) []ps.Process {
	var rv []ps.Process
	processes, _ := ps.Processes()
	for _, processe := range processes {
		if processe.Executable() == processName {
			rv = append(rv, processe)
		}
	}
	return rv
}

func readDpdkNicBindShell(filepath string) map[string]string {
	dpdkNicBindBytes, err := utils.ReadFileFast(filepath)
	utils.ErrHandlePrintln(err, "读取dpdk_nic_bind.sh失败：")
	dpdkNicBindStr := string(dpdkNicBindBytes)
	dpdkNicBindSlice := strings.Split(dpdkNicBindStr, CRLF)

	options := map[string]string{}
	for _, line := range dpdkNicBindSlice {
		if strings.Contains(line, "=") {
			words := strings.Split(line, "=")
			key := strings.TrimSpace(words[0])
			// if unicode.IsUpper(rune(key[0])) {
			if unicode.IsUpper(rune(key[0])) {
				var value string
				if len(words) >= 2 {
					value = strings.TrimSpace(strings.Trim(words[1], `"`))
					options[key] = value
				}
			}
		}
	}
	return options
}
