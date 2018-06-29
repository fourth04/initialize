// Package main provides ...
package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/Unknwon/goconfig"
	"github.com/fourth04/initialize/utils"
	"github.com/manifoldco/promptui"
)

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

func validateYesNoQuit(input string) error {
	input = strings.ToLower(input)
	ok := utils.Contains([]string{"yes", "no", "quit"}, input)
	if !ok {
		return errors.New("输入参数错误")
	}
	return nil
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

	log.Println("正在获取dpdk网卡绑定情况...")
	ifsBinded := getDpdkDevBind()
	if len(ifsBinded) != 0 {
		promptYesNoQuit := promptui.Prompt{
			Label:    fmt.Sprintf("dpdk已有绑定网卡%s，请选择是否需要解绑：yes/no/quit", strings.Join(ifsBinded, ",")),
			Validate: validateYesNoQuit,
		}
		resultDpdkNicConfig, err := promptYesNoQuit.Run()
		utils.ErrHandleFatalln(err, "输入参数错误！")
		switch resultDpdkNicConfig {
		case "yes":
			stdout, stderr, err := utils.ExecuteAndGetResult("dpdk-nic-unbind.sh")
			if stderr != "" {
				log.Fatalln("dpdk网卡解绑失败！")
			}
			utils.ErrHandleFatalln(err, "dpdk网卡解绑失败！")
			fmt.Println("执行命令：", stdout)
			log.Println("dpdk网卡解绑完成！")
		case "quit":
			os.Exit(0)
		}
	}
	log.Println("dpdk网卡未绑定！")

	log.Println("正在读取dpdk_nic_bind.sh...")
	// options := readDpdkNicBindShell("..\\docs\\dpdk_nic_bind.sh")
	options := readDpdkNicBindShell("/bin/dpdk-nic-bind.sh")
	log.Println("读取dpdk_nic_bind.sh完成！")

	dpdkNicConfigFilepath, ok := options["DPDK_NICCONF_FILE"]
	if !ok {
		log.Fatal("DPDK_NICCONF_FILE参数配置错误！")
	}
	progConfigFilepath, ok := options["PROG_CONF_FILE"]
	if !ok {
		log.Fatal("PROG_CONF_FILE参数配置错误！")
	}

	log.Println("正在检测dpdk_nic_config文件状态！")
	isFileExist, err := utils.IsFileExist(dpdkNicConfigFilepath)
	if isFileExist {
		promptYesNoQuit := promptui.Prompt{
			Label:    "检测到dpdk_nic_config已存在，请选择是否删除该文件：yes/no/quit",
			Validate: validateYesNoQuit,
		}
		resultDpdkNicConfig, err := promptYesNoQuit.Run()
		utils.ErrHandleFatalln(err, "输入参数错误！")
		switch resultDpdkNicConfig {
		case "yes":
			err := os.Remove("dpdk_nic_config")
			utils.ErrHandleFatalln(err, "删除dpdk_nic_config遇错！")
			log.Println("删除dpdk_nic_config成功！")
		case "quit":
			os.Exit(0)
		}
	}
	utils.ErrHandleFatalln(err, "检测dpdk_nic_config失败！")
	log.Println("检测到dpdk_nic_config未存在！")

	log.Println("正在获取网卡配置情况...")
	adapterNames := []string{}
	adapters := getIfInfo()
	log.Println("共读取到网卡如下：")
	for i, adapter := range adapters {
		if !strings.HasPrefix(adapter.name, "v") {
			log.Printf("[%d]:%s\n", i, adapter.name)
			adapterNames = append(adapterNames, adapter.name)
		}
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
		Label:    "请输入所选网卡序号，多选请用逗号分隔",
		Validate: validate,
	}

	inputIfs, err := prompt.Run()
	utils.ErrHandleFatalln(err, "网卡选择错误，中断网卡绑定操作！")

	indexsIfs := strings.Split(inputIfs, ",")
	ifsSelectedSlice := []string{}
	for _, index := range indexsIfs {
		i, _ := strconv.Atoi(index)
		ifsSelectedSlice = append(ifsSelectedSlice, adapterNames[i])
	}
	ifsSelectedStr := strings.Join(ifsSelectedSlice, ",")
	log.Printf("你选择了网卡：%s\n", ifsSelectedStr)

	log.Println("正在加载sdpi.ini配置文件...")
	setIniFile(progConfigFilepath, "dns", "in_nic", ifsSelectedStr)
	log.Println("修改sdpi.ini配置文件完成！")

	log.Println("正在尝试进行dpdk网卡绑定...")
	stdout, stderr, err := utils.ExecuteAndGetResult("dpdk-nic-bind.sh")
	if stderr != "" {
		log.Fatalln("dpdk网卡绑定失败！")
	}
	utils.ErrHandleFatalln(err, "dpdk网卡绑定失败！")
	fmt.Println("执行命令：", stdout)
	log.Println("dpdk网卡绑定完成！")

	log.Println("正在检测dpdk_nic_config文件生成状态！")
	isFileExist, err = utils.IsFileExist(dpdkNicConfigFilepath)
	if !isFileExist {
		log.Println("生成dpdk_nic_config文件失败！")
	}
	utils.ErrHandleFatalln(err, "检测到dpdk_nic_config生成失败！")
	log.Println("dpdk_nic_config文件生成成功！")

	log.Println("正在尝试进程启动...")
	stdout, stderr, err = utils.ExecuteAndGetResult("service sdpid restart")
	if stderr != "" {
		log.Fatalln("进程启动失败！")
	}
	utils.ErrHandleFatalln(err, "进程启动失败！")
	fmt.Println("执行命令：", stdout)
	log.Println("进程启动完成！")

	log.Println("正在检测进程运行情况...")
	// processes := getProcesses("sdip start")
	processes := getProcesses("python.exe")
	if len(processes) == 0 {
		log.Fatalln("进程启动失败，请检测失败原因！")
	}
	log.Println("进程运行正常！")

	vIfsSelectedStr := make([]string, len(ifsSelectedSlice))
	for ix, ifStr := range ifsSelectedSlice {
		vIfsSelectedStr[ix] = "v" + ifStr
	}

	/* for {
		log.Println("虚拟网卡启动中，请耐心等待30秒...")
		count := 0
		time.Sleep(time.Second * 30)
		log.Println("正在检测dpdk虚拟网卡启动情况...")
		adapters := getIfInfo()
		for _, adapter := range adapters {
			ok := utils.Contains(vIfsSelectedStr, adapter.name)
			if ok {
				log.Printf("%s启动成功！\n", adapter.name)
				count += 1
			}
		}
		if count == len(vIfsSelectedStr) {
			log.Println("dpdk虚拟网卡启动完成！")
			break
		}
	} */

	validateIP := func(input string) error {
		ip := net.ParseIP(input)
		if ip == nil {
			return errors.New("输入格式错误！")
		}
		return nil
	}

	log.Println("开始进行虚拟网卡配置...")
	for _, ifStr := range vIfsSelectedStr {
		promptIP := promptui.Prompt{
			Label:    "请输入" + ifStr + "的IP地址",
			Validate: validateIP,
		}

		resultIP, err := promptIP.Run()
		utils.ErrHandleFatalln(err, "输入参数错误")

		promptNetmask := promptui.Prompt{
			Label:    "请输入" + ifStr + "的子网掩码",
			Validate: validateIP,
		}
		resultNetmask, err := promptNetmask.Run()
		utils.ErrHandleFatalln(err, "输入参数错误")

		stdout, stderr, err := utils.ExecuteAndGetResult(fmt.Sprintf("ifconfig %s %s netmask %s", ifStr, resultIP, resultNetmask))
		if stderr != "" {
			log.Fatalln(ifStr + "配置失败！")
		}
		utils.ErrHandleFatalln(err, ifStr+"配置失败！")
		fmt.Println("执行命令：", stdout)
	}
	log.Println("虚拟网卡配置完成！")

}
