package utils

import (
	"errors"
	"fmt"
	"net"
	"regexp"
	"strconv"
)

// CIDRMask2IPMask converts CIDRMask to IPMask
func CIDRMask2IPMask(length int) string {
	IPMask := net.CIDRMask(length, 32)
	return fmt.Sprintf("%d.%d.%d.%d", IPMask[0], IPMask[1], IPMask[2], IPMask[3])
}

// IPMask2CIDRMask converts IPMask to CIDRMask
func IPMask2CIDRMask(ip string) int {
	mask := net.IPMask(net.ParseIP(ip).To4()) // If you have the mask as a string
	//mask := net.IPv4Mask(255,255,255,0) // If you have the mask as 4 integer values

	prefixSize, _ := mask.Size()
	return prefixSize
}

// MaskConvert convert mask format
func MaskConvert(input string) (string, string, error) {
	inputLength := len(input)
	fmt.Println(inputLength)
	switch {
	case inputLength >= 1 && inputLength <= 2:
		length, err := strconv.Atoi(input)
		if err != nil {
			return "", "", errors.New("input format error")
		}
		if length <= 0 || length > 32 {
			return "", "", errors.New("input format error")
		}
		return CIDRMask2IPMask(length), input, nil
	case inputLength >= 7 && inputLength <= 15:
		pattern := regexp.MustCompile(`^(?:\d+\.){3}\d+$`)
		if !pattern.MatchString(input) {
			return "", "", errors.New("input format error")
		}
		CIDRMask := IPMask2CIDRMask(input)
		if CIDRMask == 0 {
			return "", "", errors.New("input format error")
		}
		return input, strconv.Itoa(CIDRMask), nil
	default:
		return "", "", errors.New("input format error")
	}
}
