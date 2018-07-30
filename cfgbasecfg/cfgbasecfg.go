package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/fourth04/initialize/sysinfo"
	"github.com/fourth04/initialize/utils"
	"github.com/manifoldco/promptui"
)

func main() {
	// ==================== 第一步，配置Linux系统相关配置 ====================
	log.Println("正在配置系统hostname...")
	hostname, _ := utils.ExecuteAndGetResultCombineError("hostname")
	promptHostname := promptui.Prompt{
		Label: fmt.Sprintf("当前系统hostname为%s：，请输入新hostname，直接回车则跳过", hostname),
	}
	newHostname, err := promptHostname.Run()
	utils.ErrHandleFatalln(err, "输入参数错误！")
	if newHostname != "" {
		_, err := utils.ExecuteAndGetResultCombineError(fmt.Sprintf("hostname %s", newHostname))
		utils.ErrHandleFatalln(err, "更改hostname失败！")
	}
	log.Println("配置系统hostname完成！")

	log.Println("正在配置NTP服务地址...")
	ntpIP, err := sysinfo.GetNtpIP()
	utils.ErrHandleFatalln(err, "获取NTP服务地址失败！")
	var promptNtpIPLable string
	if ntpIP == "" {
		promptNtpIPLable = fmt.Sprintf("当前未配置NTP服务地址，请输入新地址，直接回车则跳过")
	} else {
		promptNtpIPLable = fmt.Sprintf("当前配置的NTP服务地址为：%s，请输入新地址，直接回车则跳过", ntpIP)
	}
	promptNtpIP := promptui.Prompt{
		Label:    promptNtpIPLable,
		Validate: utils.ValidateIPOrNil,
	}
	newNtpIP, err := promptNtpIP.Run()
	utils.ErrHandleFatalln(err, "输入参数错误！")
	if newNtpIP != "" {
		err := sysinfo.CfgNtpIP(newNtpIP)
		utils.ErrHandleFatalln(err, "更新NTP crontab失败！")
	}
	log.Println("配置NTP服务地址完成！")

	log.Println("正在配置MS服务地址...")
	msAgentIniCfg, err := sysinfo.GetMsAgentIniCfg()
	utils.ErrHandleFatalln(err, "获取MS服务地址失败！")
	promptMsIP := promptui.Prompt{
		Label:    fmt.Sprintf("当前MS服务地址为%s，请输入新地址，直接回车则跳过", msAgentIniCfg["ms_host"]),
		Validate: utils.ValidateIPOrNil,
	}
	newMsIP, err := promptMsIP.Run()
	utils.ErrHandleFatalln(err, "输入参数错误！")
	if newMsIP != "" {
		err = sysinfo.SetIniFile("/etc/msagent.ini", "ms", "host", newMsIP)
		utils.ErrHandleFatalln(err, "更新MS服务地址失败！")
	}
	log.Println("配置MS服务地址完成！")

	log.Println("正在配置DNS服务地址...")
	dnsIPs, err := sysinfo.GetDnsIPs()
	utils.ErrHandleFatalln(err, "获取DNS服务地址失败！")
	var promptDnsIPsLable string
	if len(dnsIPs) == 0 {
		promptDnsIPsLable = fmt.Sprintf("当前未配置DNS服务地址，请输入新地址，直接回车则跳过")
	} else {
		promptDnsIPsLable = fmt.Sprintf("当前配置的DNS服务地址为：%s，请输入新地址，直接回车则跳过", strings.Join(dnsIPs, ","))
	}
	promptDnsIPs := promptui.Prompt{
		Label:    promptDnsIPsLable,
		Validate: utils.ValidateIPsOrNil,
	}
	newDnsIPs, err := promptDnsIPs.Run()
	utils.ErrHandleFatalln(err, "输入参数错误！")
	if newDnsIPs != "" {
		ips := strings.Split(newDnsIPs, ",")
		err := sysinfo.CfgDnsIPs(ips)
		utils.ErrHandleFatalln(err, "更新DNS服务地址失败！")
	}
	log.Println("配置DNS服务地址完成！")
}
