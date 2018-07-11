package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Unknwon/goconfig"
	"github.com/fourth04/initialize/sysinfo"
	"github.com/fourth04/initialize/utils"
	"github.com/manifoldco/promptui"
)

var MANAGERIP = "192.168.128.114"

func initLog() {
	log.SetPrefix("[INFO]")
}

func init() {
	initLog()
}

func setIniFile(filepath, section, key, value string) {
	cfg, err := goconfig.LoadConfigFile(filepath)
	utils.ErrHandleFatalln(err, "加载sdpi.ini配置文件失败：")
	cfg.SetValue(section, key, value)
	err = goconfig.SaveConfigFile(cfg, filepath)
	utils.ErrHandlePrintln(err, "保存sdpi.ini配置文件失败：")
}

func validateYesNo(input string) error {
	input = strings.ToLower(input)
	ok := utils.Contains([]string{"yes", "no"}, input)
	if !ok {
		return errors.New("请输入yes/no！")
	}
	return nil
}

func validateIP(input string) error {
	ip := net.ParseIP(input)
	if ip == nil {
		return errors.New("请输入正确IP格式！")
	}
	return nil
}

func validateIPOrNil(input string) error {
	if input == "" {
		return nil
	}
	ip := net.ParseIP(input)
	if ip == nil {
		return errors.New("请输入正确IP格式，或者直接回车跳过！")
	}
	return nil
}

func main() {
	/*
		==================== 第零步，状态检测 ====================
		1. dpdk驱动安装状态，主要检查是否未执行dpdk-setup.sh来安装dpdk驱动
		2. 主进程运行状态，主要检测是否主进程未退出
		3. dpdk网卡绑定状态，主要检测是否dpdk已经绑定了网卡
		4. dpdk_nic_config文件生成状态，主要检测是否有dpdk_nic_config未删除
	*/
	log.Println("正在获取dpdk驱动安装情况...")
	IsDpdkDriverOKFlag := sysinfo.IsDpdkDriverOK()
	if !IsDpdkDriverOKFlag {
		promptYesNoQuit := promptui.Prompt{
			Label:    "dpdk驱动未安装，请选择是否需要安装驱动：yes/no",
			Validate: validateYesNo,
		}
		resultDpdkNicConfig, err := promptYesNoQuit.Run()
		utils.ErrHandleFatalln(err, "输入参数错误！")
		switch resultDpdkNicConfig {
		case "yes":
			_, err := utils.ExecuteAndGetResultCombineError("dpdk-setup.sh")
			utils.ErrHandleFatalln(err, "dpdk驱动安装失败！")
			log.Println("dpdk驱动安装成功！")
		case "no":
			os.Exit(0)
		}
	}
	log.Println("dpdk驱动已安装！")

	log.Println("正在检测进程运行情况...")
	IsProcessRuningFlag := sysinfo.IsProcessRuning("sdpi")
	if IsProcessRuningFlag {
		promptYesNoQuit := promptui.Prompt{
			Label:    "进程已启动，请选择是否需要停止进程：yes/no",
			Validate: validateYesNo,
		}
		resultDpdkNicConfig, err := promptYesNoQuit.Run()
		utils.ErrHandleFatalln(err, "输入参数错误！")
		switch resultDpdkNicConfig {
		case "yes":
			_, err := utils.ExecuteAndGetResultCombineError("service sdpid stop")
			utils.ErrHandleFatalln(err, "进程停止失败！")
			log.Println("进程停止成功！")
		case "no":
			os.Exit(0)
		}
	}
	log.Println("进程未运行！")

	log.Println("正在获取dpdk网卡绑定情况...")
	ifsBinded := sysinfo.GetDpdkDevBind()
	if len(ifsBinded) != 0 {
		promptYesNoQuit := promptui.Prompt{
			Label:    fmt.Sprintf("dpdk已有绑定网卡%s，请选择是否需要解绑：yes/no", strings.Join(ifsBinded, ",")),
			Validate: validateYesNo,
		}
		resultDpdkNicConfig, err := promptYesNoQuit.Run()
		utils.ErrHandleFatalln(err, "输入参数错误！")
		switch resultDpdkNicConfig {
		case "yes":
			_, err := utils.ExecuteAndGetResultCombineError("dpdk-nic-unbind.sh")
			utils.ErrHandleFatalln(err, "dpdk网卡解绑失败！")
			log.Println("dpdk网卡解绑完成！")
		case "no":
			os.Exit(0)
		}
	}
	log.Println("dpdk网卡未绑定！")

	log.Println("正在读取dpdk_nic_bind.sh...")
	options, err := sysinfo.ReadDpdkNicBindShell("/bin/dpdk-nic-bind.sh")
	utils.ErrHandleFatalln(err, "读取dpdk_nic_bind.sh失败：")

	dpdkNicConfigFilepath, ok := options["DPDK_NICCONF_FILE"]
	if !ok {
		log.Fatal("dpdk_nic_bind.sh DPDK_NICCONF_FILE参数配置错误！")
	}
	progConfigFilepath, ok := options["PROG_CONF_FILE"]
	if !ok {
		log.Fatal("dpdk_nic_bind.sh PROG_CONF_FILE参数配置错误！")
	}
	log.Println("读取dpdk_nic_bind.sh完成！")

	log.Println("正在检测dpdk_nic_config文件状态！")

	isDpdkNicConfigFileExist, err := utils.IsFileExist(dpdkNicConfigFilepath)
	if isDpdkNicConfigFileExist {
		promptYesNoQuit := promptui.Prompt{
			Label:    "检测到dpdk_nic_config已存在，请选择是否删除该文件：yes/no",
			Validate: validateYesNo,
		}
		resultDpdkNicConfig, err := promptYesNoQuit.Run()
		utils.ErrHandleFatalln(err, "输入参数错误！")
		switch resultDpdkNicConfig {
		case "yes":
			err := os.Remove("dpdk_nic_config")
			utils.ErrHandleFatalln(err, "删除dpdk_nic_config遇错！")
			log.Println("删除dpdk_nic_config成功！")
		case "no":
			os.Exit(0)
		}
	}
	utils.ErrHandleFatalln(err, "检测dpdk_nic_config失败！")
	log.Println("检测到dpdk_nic_config未存在！")

	// ==================== 第一步，配置Linux系统相关配置 ====================

	// ==================== 第二步，管理网卡配置 ====================
	promptYesNoQuit := promptui.Prompt{
		Label:    "请确认是否需要自定义管理网卡：yes/no",
		Validate: validateYesNo,
	}
	resultAdminPort, err := promptYesNoQuit.Run()
	utils.ErrHandleFatalln(err, "输入参数错误！")
	switch resultAdminPort {
	case "yes":
		log.Println("正在获取管理网卡配置信息...")
		adapter := sysinfo.GetIfInfoManage(MANAGERIP)
		if adapter != nil {
			log.Println("获取到管理网卡信息如下：")
			log.Printf("网卡:%s IPv4地址:%s/%d IPv6地址:%s/%d\n", adapter.Name, adapter.Inet, adapter.Netmasklen, adapter.Inet6, adapter.Prefixlen)

			ifStr := adapter.Name + ":1"

			promptIP := promptui.Prompt{
				Label:    "请输入" + ifStr + "的IP地址",
				Validate: validateIP,
			}

			resultIP, err := promptIP.Run()
			utils.ErrHandleFatalln(err, "输入参数错误！")

			promptNetmask := promptui.Prompt{
				Label:    "请输入" + ifStr + "的子网掩码",
				Validate: validateIP,
			}
			resultNetmask, err := promptNetmask.Run()
			utils.ErrHandleFatalln(err, "输入参数错误！")

			ifcfg := sysinfo.NewDefaultIfCfg()
			ifcfg.DEVICE = ifStr
			ifcfg.IPADDR = resultIP
			ifcfg.NETMASK = resultNetmask

			err = ifcfg.SaveConfigFile("/etc/sysconfig/network-scripts/ifcfg-" + ifcfg.DEVICE)
			utils.ErrHandleFatalln(err, "管理网卡配置文件生成失败！")
			log.Println("管理网卡配置文件生成成功！")
		} else {
			log.Println("未获取管理网卡的信息!")
		}
	}

	// ==================== 第三步，业务网卡配置 ====================
	log.Println("正在获取网卡配置情况...")
	adapterNames := []string{}
	adapters := sysinfo.GetIfInfoService(MANAGERIP)
	log.Println("共读取到网卡如下：")
	for i, adapter := range adapters {
		log.Printf("[%d] 网卡:%s IPv4地址:%s/%d IPv6地址:%s/%d\n", i, adapter.Name, adapter.Inet, adapter.Netmasklen, adapter.Inet6, adapter.Prefixlen)
		adapterNames = append(adapterNames, adapter.Name)
	}
	log.Println("获取网卡配置情况完成！")

	validate := func(input string) error {
		indexsIfs := strings.Split(input, ",")
		for _, index := range indexsIfs {
			i, err := strconv.Atoi(index)
			if err != nil {
				return errors.New(fmt.Sprintf("%s序号输入错误\n", index))
			}
			if i > len(adapterNames) {
				return errors.New(fmt.Sprintf("%s序号输入错误\n", index))
			}
		}
		return nil
	}
	prompt := promptui.Prompt{
		Label:    "请输入需绑定网卡序号，多选请用逗号分隔",
		Validate: validate,
	}

	inputIfs, err := prompt.Run()
	utils.ErrHandleFatalln(err, "网卡选择错误，中断网卡绑定操作！")

	indexsIfs := strings.Split(inputIfs, ",")
	ifsSelected := []string{}
	for _, index := range indexsIfs {
		i, _ := strconv.Atoi(index)
		ifsSelected = append(ifsSelected, adapterNames[i])
	}
	ifsSelectedStr := strings.Join(ifsSelected, ",")
	log.Printf("你选择了网卡：%s\n", ifsSelectedStr)

	log.Println("正在修改sdpi.ini配置文件...")
	setIniFile(progConfigFilepath, "dns", "in_nic", ifsSelectedStr)
	log.Println("修改sdpi.ini配置文件完成！")

	log.Println("正在尝试进行dpdk网卡绑定...")
	_, err = utils.ExecuteAndGetResultCombineError("dpdk-nic-bind.sh")
	utils.ErrHandleFatalln(err, "dpdk网卡绑定失败！")
	log.Println("dpdk网卡绑定完成！")

	log.Println("正在检测dpdk_nic_config文件生成状态！")
	isDpdkNicConfigFileExist, err = utils.IsFileExist(dpdkNicConfigFilepath)
	if !isDpdkNicConfigFileExist {
		log.Println("生成dpdk_nic_config文件失败！")
	}
	utils.ErrHandleFatalln(err, "检测到dpdk_nic_config生成失败！")
	log.Println("dpdk_nic_config文件生成成功！")

	log.Println("正在尝试进程启动...")
	_, err = utils.ExecuteAndGetResultCombineError("service sdpid restart")
	utils.ErrHandleFatalln(err, "进程启动失败！")
	log.Println("进程启动完成！")

	log.Println("正在检测进程运行情况...")
	processes := sysinfo.GetProcesses("sdpi")
	if len(processes) == 0 {
		log.Fatalln("进程启动失败，请检测失败原因！")
	}
	log.Println("进程运行正常！")

	vIfsSelected := make([]string, len(ifsSelected))
	for ix, ifStr := range ifsSelected {
		vIfsSelected[ix] = "v" + ifStr
	}

	N := 6
	for i := 0; i < N; i++ {
		log.Println("虚拟网卡启动中，请耐心等待30秒...")
		count := 0
		time.Sleep(time.Second * 30)
		log.Println("正在检测dpdk虚拟网卡启动情况...")
		adapters := sysinfo.GetIfInfo()
		for _, adapter := range adapters {
			ok := utils.Contains(vIfsSelected, adapter.Name)
			if ok {
				log.Printf("%s启动成功！\n", adapter.Name)
				count += 1
			}
		}
		if count == len(vIfsSelected) {
			log.Println("dpdk虚拟网卡启动完成！")
			break
		}
		if i == N {
			log.Fatalf("已等待%d秒，虚拟网卡仍未启动，请检查！\n", 30*N)
		}
	}

	log.Println("开始进行业务网卡配置...")
	var vIfsSelectedIfCfgs []sysinfo.IfCfg
	lengthVIfsSelected := len(vIfsSelected)
	for _, ifStr := range vIfsSelected {
		promptIP := promptui.Prompt{
			Label:    "请输入" + ifStr + "的IP地址",
			Validate: validateIP,
		}

		resultIP, err := promptIP.Run()
		utils.ErrHandleFatalln(err, "输入参数错误！")

		promptNetmask := promptui.Prompt{
			Label:    "请输入" + ifStr + "的子网掩码",
			Validate: validateIP,
		}
		resultNetmask, err := promptNetmask.Run()
		utils.ErrHandleFatalln(err, "输入参数错误！")

		ifcfg := sysinfo.NewDefaultIfCfg()
		ifcfg.DEVICE = ifStr
		ifcfg.IPADDR = resultIP
		ifcfg.NETMASK = resultNetmask

		vIfsSelectedIfCfgs = append(vIfsSelectedIfCfgs, ifcfg)

		err = ifcfg.SaveConfigFile("/etc/sysconfig/network-scripts/ifcfg-" + ifcfg.DEVICE)
		utils.ErrHandleFatalln(err, "业务网卡"+ifStr+"配置文件生成失败！")
		log.Println("业务网卡" + ifStr + "配置文件生成成功！")
	}
	log.Println("业务网卡配置完成！")

	if lengthVIfsSelected == 1 {
		log.Println("开始进行业务网关配置...")
		promptGW := promptui.Prompt{
			Label:    "请输入业务网关地址，可跳过",
			Validate: validateIPOrNil,
		}
		resultGW, err := promptGW.Run()
		utils.ErrHandleFatalln(err, "输入参数错误！")

		if resultGW != "" {
			_, err = utils.ExecuteAndGetResultCombineError("route add default gw " + resultGW)
			utils.ErrHandleFatalln(err, "业务网关配置失败！")
			log.Println("业务网关配置完成！")
		}
	}

	// ==================== 最后一步，重启网络服务 ====================
	log.Println("正在进行业务网卡配置测试...")

	N = 3
	for i := 0; i < N; i++ {
		log.Println("正在重启网络服务！")
		_, err = utils.ExecuteAndGetResultCombineError("service network restart")
		utils.ErrHandleFatalln(err, "重启网络服务失败！")
		log.Println("重启网络服务成功！")

		for _, vIfcfg := range vIfsSelectedIfCfgs {
			log.Printf("正在ping测%s的IP地址%s\n", vIfcfg.DEVICE, vIfcfg.IPADDR)
			isOK, _ := sysinfo.PingDial(vIfcfg.IPADDR, time.Second*10)
			if !isOK {
				log.Println("ping测失败！")
				if i == N-1 {
					log.Fatalln("ping测失败，请检查网卡无法配置IP的原因！")
				} else {
					continue
				}
			}
			log.Printf("%sIP配置成功！\n", vIfcfg.DEVICE)
		}
	}
	log.Println("业务网卡配置成功！")

	log.Println("初始化结束！")
}
