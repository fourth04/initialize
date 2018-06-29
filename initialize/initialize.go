// Package main provides ...
package main

import (
	"log"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"unicode"

	"github.com/Unknwon/goconfig"
	prompt "github.com/c-bata/go-prompt"
	"github.com/fourth04/initialize/utils"
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

func ifconfig() []*Adapter {
	bytes, _ := utils.ReadFileFast("..\\docs\\ifconfig.txt")
	stdout := string(bytes)
	// stdout, _, _ := utils.ExecuteAndGetResult("cat ..\\docs\\ifconfig.txt")
	// log.Println(stdout)

	var pat string
	var crlf string
	switch runtime.GOOS {
	case "windows":
		crlf = "\r\n"
	case "linux":
		crlf = "\n"
	default:
		crlf = "\n"
	}

	var rv []*Adapter
	pat = "(?m)" + crlf + crlf
	reg, _ := regexp.Compile(pat)
	splitedSlice := reg.Split(stdout, -1)
	// var adapterSlice []Adapter
	for _, ifInfo := range splitedSlice {
		adapter := Adapter{}
		ifInfoSplited := strings.Split(ifInfo, crlf)

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

func readDpdkNicBindShell(filepath string) map[string]string {
	dpdkNicBindBytes, err := utils.ReadFileFast(filepath)
	utils.ErrHandlePrintln(err, "读取dpdk_nic_bind.sh失败：")
	dpdkNicBindStr := string(dpdkNicBindBytes)
	dpdkNicBindSlice := strings.Split(dpdkNicBindStr, "\r\n")

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
	// log.Println(options)
	return options
}

func getCompleter(s []prompt.Suggest) func(in prompt.Document) []prompt.Suggest {
	completer := func(in prompt.Document) []prompt.Suggest {
		return prompt.FilterHasPrefix(s, in.GetWordBeforeCursor(), true)
	}
	return completer
}

func setIniFile(filepath, section, key, value string) {
	cfg, err := goconfig.LoadConfigFile(filepath)
	utils.ErrHandleFatalln(err, "加载sdpi.ini配置文件失败：")
	cfg.SetValue(section, key, value)
	err = goconfig.SaveConfigFile(cfg, filepath)
	utils.ErrHandlePrintln(err, "保存sdpi.ini配置文件失败：")
}

func main() {
	log.Println("正在获取dpdk驱动安装情况...")
	stdout, _, _ := utils.ExecuteAndGetResult("ls /mnt")
	if !strings.Contains(stdout, "huge") {
		log.Println("dpdk驱动未安装，正在启动...")
		err := utils.Executor("dpdk-setup.sh")
		// utils.ErrHandleFatalln(err, "dpdk驱动安装失败！")
		utils.ErrHandlePrintln(err, "dpdk驱动安装失败！")
	}
	log.Println("dpdk驱动已安装！")

	log.Println("正在获取网卡配置情况...")
	s := []prompt.Suggest{}
	adapterNames := []string{}
	adapters := ifconfig()
	log.Println("共读取到网卡如下：")
	for i, adapter := range adapters {
		if !strings.HasPrefix(adapter.name, "v") {
			s = append(s, prompt.Suggest{Text: strconv.Itoa(i), Description: adapter.name})
			log.Printf("[%d]:%s\n", i, adapter.name)
			adapterNames = append(adapterNames, adapter.name)
		}
	}
	log.Println("获取网卡配置情况完成！")
	log.Println("请选择需要绑定dpdk的网卡：")

	inIfs := prompt.Input(">>> ", getCompleter(s),
		prompt.OptionTitle("sql-prompt"),
		prompt.OptionHistory([]string{"SELECT * FROM users;"}),
		prompt.OptionPrefixTextColor(prompt.Yellow),
		prompt.OptionPreviewSuggestionTextColor(prompt.Blue),
		prompt.OptionSelectedSuggestionBGColor(prompt.LightGray),
		prompt.OptionSuggestionBGColor(prompt.DarkGray))
	indexsIfs := strings.Split(inIfs, ",")
	ifsStrSlice := []string{}
	for _, index := range indexsIfs {
		i, err := strconv.Atoi(index)
		if err != nil {
			log.Printf("%s序号输入错误\n", index)
			continue
		}
		if i > len(adapterNames) {
			log.Printf("%s序号输入错误\n", index)
			continue
		}
		ifsStrSlice = append(ifsStrSlice, adapterNames[i])
	}
	if len(ifsStrSlice) == 0 {
		log.Println("输入网卡错误，停止网卡绑定操作！")
		return
	}
	ifsStrFinal := strings.Join(ifsStrSlice, ",")
	log.Printf("你选择了网卡：%s\n", ifsStrFinal)

	log.Println("正在加载sdpi.ini配置文件...")
	setIniFile("..\\docs\\sdpi.ini", "dns", "in_nic", ifsStrFinal)
	log.Println("修改sdpi.ini配置文件完成！")

	/* log.Println("正在读取dpdk_nic_bind.sh...")
	options := readDpdkNicBindShell("..\\docs\\dpdk_nic_bind.sh")
	log.Println(options)
	log.Println("读取dpdk_nic_bind.sh完成！") */

}
