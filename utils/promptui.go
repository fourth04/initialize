package utils

import (
	"errors"
	"net"
	"strings"
)

func ValidateYesNo(input string) error {
	input = strings.ToLower(input)
	ok := Contains([]string{"yes", "no"}, input)
	if !ok {
		return errors.New("请输入yes/no！")
	}
	return nil
}

func ValidateIP(input string) error {
	ip := net.ParseIP(input)
	if ip == nil {
		return errors.New("请输入正确IP格式！")
	}
	return nil
}

func ValidateIPs(input string) error {
	ips := strings.Split(input, ",")
	for _, input := range ips {
		ip := net.ParseIP(input)
		if ip == nil {
			return errors.New("请输入正确IP格式！")
		}
	}
	return nil
}

func ValidateIPOrNil(input string) error {
	if input == "" {
		return nil
	}
	ip := net.ParseIP(input)
	if ip == nil {
		return errors.New("请输入正确IP格式，直接回车则跳过！")
	}
	return nil
}

func ValidateIPsOrNil(input string) error {
	if input == "" {
		return nil
	}
	ips := strings.Split(input, ",")
	for _, input := range ips {
		ip := net.ParseIP(input)
		if ip == nil {
			return errors.New("请输入正确IP格式！")
		}
	}
	return nil
}
