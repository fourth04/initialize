package sysinfo

import (
	"fmt"
	"log"
	"net"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/Unknwon/goconfig"
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
	DEVICE             string `json:"device" binding:"required"`
	ONBOOT             string `json:"onboot"`
	IPADDR             string `json:"ipaddr" binding:"required"`
	GATEWAY            string `json:"gateway"`
	NETMASK            string `json:"netmask" binding:"required"`
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
	Name       string `json:"name" binding:"required"`
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

func GetIfInfoManage(ip string) []*Adapter {
	adapters := GetIfInfoHasPrefix("enp")
	var rv []*Adapter
	for _, adapter := range adapters {
		if adapter.Inet == ip {
			rv = append(rv, adapter)
			break
		}
	}
	for _, adapter := range adapters {
		if adapter.Name == rv[0].Name+":1" {
			rv = append(rv, adapter)
			break
		}
	}
	return rv
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

func IsProcessRunning(processName string) bool {
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

func IsDpdkNicBindShellOK() (map[string]string, bool) {
	options, err := ReadDpdkNicBindShell("/bin/dpdk-nic-bind.sh")
	if err != nil {
		return nil, false
	}
	_, ok := options["PROG_CONF_FILE"]
	if !ok {
		return nil, false
	}
	return options, true
}

func IsDpdkNicConfigFileExist() bool {
	options, ok := IsDpdkNicBindShellOK()
	if !ok {
		return false
	}
	dpdkNicConfigFilepath := options["PROG_CONF_FILE"]
	isDpdkNicConfigFileExist, err := utils.IsFileExist(dpdkNicConfigFilepath)
	if err != nil {
		return false
	}
	return isDpdkNicConfigFileExist
}

type RunningStatus struct {
	IsDpdkDriverOKFlag           bool `json:"is_dpdk_driver_ok_flag" binding:"required"`
	IsProcessRunningFlag         bool `json:"is_process_running_flag" binding:"required"`
	IsDpdkBindedFlag             bool `json:"is_dpdk_binded_flag" binding:"required"`
	IsDpdkNicBindShellOKFlag     bool `json:"is_dpdk_nic_bind_shell_ok_flag" binding:"required"`
	IsDpdkNicConfigFileExistFlag bool `json:"is_dpdk_nic_config_file_exist_flag" binding:"required"`
}

func GetRunningStatus() RunningStatus {
	var isDpdkNicBindShellOKFlag, isDpdkNicConfigFileExistFlag bool
	options, ok := IsDpdkNicBindShellOK()
	if !ok {
		isDpdkNicBindShellOKFlag = false
		isDpdkNicConfigFileExistFlag = false
	} else {
		isDpdkNicBindShellOKFlag = true
		dpdkNicConfigFilepath := options["PROG_CONF_FILE"]
		tmpFlag, err := utils.IsFileExist(dpdkNicConfigFilepath)
		if err != nil {
			isDpdkNicConfigFileExistFlag = false
		} else {
			isDpdkNicConfigFileExistFlag = tmpFlag
		}
	}
	return RunningStatus{
		IsDpdkDriverOKFlag:           IsDpdkDriverOK(),
		IsProcessRunningFlag:         IsProcessRunning("sdpi"),
		IsDpdkBindedFlag:             IsDpdkBinded(),
		IsDpdkNicBindShellOKFlag:     isDpdkNicBindShellOKFlag,
		IsDpdkNicConfigFileExistFlag: isDpdkNicConfigFileExistFlag,
	}
}

func ReadDpdkNicBindShell(filepath string) (map[string]string, error) {
	dpdkNicBindSlice, err := utils.ReadFileFast2Slice(filepath)
	if err != nil {
		return nil, err
	}

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

func GetCrontabSlice() ([]string, error) {
	crontabBytes, err := utils.ReadFileFast("/var/spool/cron/root")
	if err != nil {
		return nil, err
	}
	crontabStr := string(crontabBytes)
	crontabSlice := strings.Split(crontabStr, utils.CRLF)
	return crontabSlice, nil
}

func GetNtpIP() (string, error) {
	crontabSlice, err := GetCrontabSlice()
	if err != nil {
		return "", err
	}
	var ntpIP string
	for _, line := range crontabSlice {
		if strings.Contains(line, "ntpdate") {
			words := strings.Split(line, " ")
			ntpIP = words[len(words)-1]
		}
	}
	return ntpIP, nil
}

func CfgNtpIP(ntpIP string) error {
	crontabSlice, err := GetCrontabSlice()
	if err != nil {
		return err
	}
	crontabNtp := fmt.Sprintf("*/30 * * * * ntpdate %s", ntpIP)
	isNew := true
	for ix, line := range crontabSlice {
		if strings.Contains(line, "ntpdate") {
			isNew = false
			crontabSlice[ix] = crontabNtp
		}
	}
	if isNew {
		crontabSlice = append(crontabSlice, crontabNtp)
	}
	crontabOutStr := strings.Join(crontabSlice, utils.CRLF)
	err = utils.WriteFileFast("/var/spool/cron/root", []byte(crontabOutStr))
	if err != nil {
		return err
	}
	return nil
}

func GetMsAgentIniCfg() (map[string]string, error) {
	cfg, err := goconfig.LoadConfigFile("/etc/msagent.ini")
	if err != nil {
		return nil, err
	}
	rv := make(map[string]string)
	value, err := cfg.GetValue("device", "sn")
	if err != nil {
		return nil, err
	}
	rv["device_sn"] = value
	value, err = cfg.GetValue("ms", "host")
	if err != nil {
		return nil, err
	}
	rv["ms_host"] = value
	return rv, nil
}

func SetIniFile(filepath, section, key, value string) error {
	cfg, err := goconfig.LoadConfigFile(filepath)
	if err != nil {
		return err
	}
	cfg.SetValue(section, key, value)
	err = goconfig.SaveConfigFile(cfg, filepath)
	if err != nil {
		return err
	}
	return nil
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

type Route struct {
	Destination string
	Nexthop     string
}

func GetRouteInfoManage(ifName string) ([]Route, error) {
	lines, err := utils.ReadFileFast2Slice("/etc/sysconfig/network-scripts/route-" + ifName)
	if err != nil {
		return nil, err
	}
	var rv []Route
	for _, line := range lines {
		words := strings.Split(line, "via")
		if len(words) == 2 {
			route := Route{strings.TrimSpace(words[0]), strings.TrimSpace(words[1])}
			rv = append(rv, route)
		}
	}
	return rv, nil
}

func CfgManageRoute(ifName string, routes []Route) error {
	var lines []string
	for _, route := range routes {
		lines = append(lines, fmt.Sprintf("%s via %s", route.Destination, route.Nexthop))
	}
	content := []byte(strings.Join(lines, utils.CRLF))
	err := utils.WriteFileFast("/etc/sysconfig/network-scripts/route-"+ifName, content)
	if err != nil {
		return err
	}
	return nil
}

func CfgServiceIf(device, ipaddr, netmask, saveDirpath string) (IfCfg, error) {
	ifcfg := NewDefaultIfCfg()
	ifcfg.DEVICE = device
	ifcfg.IPADDR = ipaddr
	ifcfg.NETMASK = netmask

	_, err := utils.ExecuteAndGetResultCombineError(fmt.Sprintf("ifconfig %s %s netmask %s", device, ipaddr, netmask))
	if err != nil {
		return ifcfg, err
	}

	saveFilepath := filepath.Join(saveDirpath, "ifcfg-"+ifcfg.DEVICE)
	err = ifcfg.SaveConfigFile(saveFilepath)
	if err != nil {
		return ifcfg, err
	}

	return ifcfg, err
}
