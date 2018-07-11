package sysinfo

import (
	"log"
	"net"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/fourth04/initialize/utils"
	ps "github.com/mitchellh/go-ps"
	fastping "github.com/tatsushid/go-fastping"
)

type IfCfg struct {
	TYPE               string `json:"type"`
	BOOTPROTO          string `json:"bootproto"`
	DEFROUTE           string `json:"defroute"`
	PEERDNS            string `json:"peerdns"`
	PEERROUTES         string `json:"peerroutes"`
	IPV4_FAILURE_FATAL string `json:"ipv4_failure_fatal"`
	IPV6INIT           string `json:"ipv6_init"`
	IPV6_AUTOCONF      string `json:"ipv6_autoconf"`
	IPV6_DEFROUTE      string `json:"ipv6_defroute"`
	IPV6_PEERDNS       string `json:"ipv6_peerdns"`
	IPV6_PEERROUTES    string `json:"ipv6_peerroutes"`
	IPV6_FAILURE_FATAL string `json:"ipv6_failure_fatal"`
	DEVICE             string `json:"device"`
	ONBOOT             string `json:"onboot"`
	IPADDR             string `json:"ipaddr"`
	GATEWAY            string `json:"gateway"`
	NETMASK            string `json:"netmask"`
}

func (ifcfg IfCfg) SaveConfigFile(filepath string) error {
	var cfgStr string
	t := reflect.TypeOf(ifcfg)
	v := reflect.ValueOf(ifcfg)
	for i := 0; i < t.NumField(); i++ {
		key := t.Field(i).Name
		value := v.Field(i).Interface().(string)
		if value == "" {
			cfgStr += "#" + key + "=" + value + utils.CRLF
		} else {
			cfgStr += key + "=" + value + utils.CRLF
		}
	}
	cfgStr = strings.TrimSpace(cfgStr)
	err := utils.WriteFileFast(filepath, []byte(cfgStr))
	return err
}

func NewDefaultIfCfg() IfCfg {
	return IfCfg{
		TYPE:               "Ethernet",
		BOOTPROTO:          "static",
		DEFROUTE:           "yes",
		PEERDNS:            "yes",
		PEERROUTES:         "yes",
		IPV4_FAILURE_FATAL: "no",
		IPV6INIT:           "yes",
		IPV6_AUTOCONF:      "yes",
		IPV6_DEFROUTE:      "yes",
		IPV6_PEERDNS:       "yes",
		IPV6_PEERROUTES:    "yes",
		IPV6_FAILURE_FATAL: "no",
		DEVICE:             "",
		ONBOOT:             "yes",
		IPADDR:             "",
		GATEWAY:            "",
		NETMASK:            "",
	}
}

type Adapter struct {
	Name       string
	Flags      []string
	Mtu        int
	Inet       string
	Netmask    string
	Netmasklen int
	Broadcast  string
	Inet6      string
	Prefixlen  int
	Ether      string
	Txqueuelen int
}

func GetIfInfo() []*Adapter {
	// bytes, _ := utils.ReadFileFast("..\\docs\\ifconfig.txt")
	// stdout := string(bytes)
	// stdout, _, _ := utils.ExecuteAndGetResult("cat ..\\docs\\ifconfig.txt")
	stdout, _, _ := utils.ExecuteAndGetResult("ifconfig -a")

	var pat string

	var rv []*Adapter
	pat = "(?m)" + utils.CRLF + utils.CRLF
	reg, _ := regexp.Compile(pat)
	splitedSlice := reg.Split(stdout, -1)
	for _, ifInfo := range splitedSlice {
		if strings.TrimSpace(ifInfo) == "" {
			continue
		}
		adapter := Adapter{}
		ifInfoSplited := strings.Split(ifInfo, utils.CRLF)

		for _, line := range ifInfoSplited {
			var words []string
			for _, word := range strings.Split(line, " ") {
				if word != "" {
					words = append(words, word)
				}
			}
			switch {
			case strings.Contains(line, "mtu"):
				adapter.Name = strings.Trim(words[0], ":")
				adapter.Flags = strings.Split(strings.Trim(strings.Split(words[1], "<")[1], ">"), ",")
				adapter.Mtu, _ = strconv.Atoi(words[len(words)-1])
			case strings.Contains(line, "inet6"):
				adapter.Inet6 = words[1]
				adapter.Prefixlen, _ = strconv.Atoi(words[3])
			case strings.Contains(line, "inet"):
				adapter.Inet = words[1]
				adapter.Netmask = words[3]
				if len(words) == 6 {
					adapter.Broadcast = words[5]
				}
			case strings.Contains(line, "ether"):
				adapter.Ether = words[1]
				adapter.Txqueuelen, _ = strconv.Atoi(words[3])
			}
		}
		if adapter.Netmask != "" {
			mask := net.IPMask(net.ParseIP(adapter.Netmask).To4())
			prefixSize, _ := mask.Size()
			adapter.Netmasklen = prefixSize
		}
		rv = append(rv, &adapter)
	}
	// log.Println(len(reg.Split(string(bytes), -1)))
	return rv
}

func GetIfInfoHasPrefix(prefix string) []*Adapter {
	var rv []*Adapter
	adapters := GetIfInfo()
	for _, adapter := range adapters {
		if strings.HasPrefix(adapter.Name, prefix) {
			rv = append(rv, adapter)
		}
	}
	return rv
}

func GetIfInfoManage(ip string) *Adapter {
	adapters := GetIfInfoHasPrefix("enp")
	for _, adapter := range adapters {
		if adapter.Inet == ip {
			return adapter
		}
	}
	return nil
}

func GetIfInfoService(ip string) []*Adapter {
	var rv []*Adapter
	var ifManager *Adapter
	adapters := GetIfInfoHasPrefix("enp")
	for _, adapter := range adapters {
		if adapter.Inet == ip {
			ifManager = adapter
		}
	}
	for _, adapter := range adapters {
		if !strings.HasPrefix(adapter.Name, ifManager.Name) {
			rv = append(rv, adapter)
		}
	}
	return rv
}

func GetDpdkDevBind() []string {
	stdout, stderr, err := utils.ExecuteAndGetResult("dpdk-devbind.py --status | grep 'drv=igb_uio' | awk '{print $1}'")
	if stderr != "" || err != nil {
		log.Println("获取dpdk网卡绑定状态失败！", stderr)
		utils.ErrHandlePrintln(err, "获取dpdk网卡绑定状态失败！")
		return []string{}
	}
	log.Println("执行结果：", stdout)
	return strings.Split(stdout, "\n")
}

func GetProcesses(processName string) []ps.Process {
	var rv []ps.Process
	processes, _ := ps.Processes()
	for _, process := range processes {
		// fmt.Println(process)
		if process.Executable() == processName {
			rv = append(rv, process)
		}
	}
	return rv
}

func IsDpdkDriverOK() bool {
	log.Println("正在获取dpdk驱动安装情况...")
	stdout, _, _ := utils.ExecuteAndGetResult("ls /mnt")
	if !strings.Contains(stdout, "huge") {
		return false
	}
	return true
}

func IsProcessRuning(processName string) bool {
	processes := GetProcesses(processName)
	if len(processes) == 0 {
		return false
	}
	return true
}

func IsDpdkBinded() bool {
	ifsBinded := GetDpdkDevBind()
	if len(ifsBinded) == 0 {
		return false
	}
	return true
}

func IsDpdkNicConfigFileExist() bool {
	options, err := ReadDpdkNicBindShell("/bin/dpdk-nic-bind.sh")
	if err != nil {
		return false
	}
	dpdkNicConfigFilepath, ok := options["PROG_CONF_FILE"]
	if !ok {
		return false
	}
	isDpdkNicConfigFileExist, err := utils.IsFileExist(dpdkNicConfigFilepath)
	if err != nil {
		return false
	}
	return isDpdkNicConfigFileExist
}

func GetRuningStatus() map[string]bool {
	return map[string]bool{
		"IsDpdkDriverOK":           IsDpdkDriverOK(),
		"IsProcessRuning":          IsProcessRuning("sdpi"),
		"IsDpdkBinded":             IsDpdkBinded(),
		"IsDpdkNicConfigFileExist": IsDpdkNicConfigFileExist(),
	}
}

func ReadDpdkNicBindShell(filepath string) (map[string]string, error) {
	dpdkNicBindBytes, err := utils.ReadFileFast(filepath)
	if err != nil {
		return nil, err
	}
	dpdkNicBindStr := string(dpdkNicBindBytes)
	dpdkNicBindSlice := strings.Split(dpdkNicBindStr, utils.CRLF)

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
	return options, nil
}

func PingDial(ipAddr string, timeout time.Duration) (bool, error) {
	p := fastping.NewPinger()
	ra, err := net.ResolveIPAddr("ip4:icmp", ipAddr)
	if err != nil {
		return false, err
	}
	p.AddIPAddr(ra)
	p.MaxRTT = timeout

	chOnRecv := make(chan bool)
	chOnIdle := make(chan bool)

	p.OnRecv = func(addr *net.IPAddr, t time.Duration) {
		chOnRecv <- true
	}
	p.OnIdle = func() {
		chOnIdle <- false
	}
	p.RunLoop()

	var result bool
	select {
	case res := <-chOnRecv:
		result = res
	case res := <-chOnIdle:
		result = res
	case <-p.Done():
		if err = p.Err(); err != nil {
			return false, err
		}
		result = false
	}
	p.Stop()
	return result, nil
}
