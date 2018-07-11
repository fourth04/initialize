package utils

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"strings"
	"time"
)

// FormatDate 格式化时间
func FormatDate(t time.Time) string {
	year, month, day := t.Date()
	return fmt.Sprintf("%d-%02d-%02d", year, month, day)
}

// Encrypt is encrypt the data with salt
func Encrypt(data string, salt string) (string, error) {
	hash := md5.New()
	_, err := hash.Write([]byte(salt))
	if err != nil {
		return "", err
	}
	cipher := hash.Sum(nil)

	buf := new(bytes.Buffer)
	buf.Write(cipher)
	buf.WriteString(data)
	_, err = hash.Write(buf.Bytes())
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// StructToMap will change struct to map
func StructToMap(inter interface{}) map[string]interface{} {
	param := make(map[string]interface{})

	t := reflect.TypeOf(inter)
	v := reflect.ValueOf(inter)
	for i := 0; i < t.NumField(); i++ {
		param[t.Field(i).Name] = v.Field(i).Interface()
	}
	return param
}

// RandomString generate random string
func RandomString(length int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < length; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}

// Contains slice contain sub
func Contains(slice []string, sub string) bool {
	for _, str := range slice {
		if str == sub {
			return true
		}
	}
	return false
}

// ReadFileFast read file quickly
func ReadFileFast(filepath string) ([]byte, error) {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Printf("An error occurred on opening the inputfile\n" +
			"Does the file exist?\n" +
			"Have you got acces to it?\n")
		return []byte{}, err
	}
	return data, nil
}

// WriteFileFast write file quickly
func WriteFileFast(filepath string, content []byte) error {
	err := ioutil.WriteFile(filepath, content, 0666)
	if err != nil {
		return err
	}
	return nil
}

// ErrHandlePrintln log error
func ErrHandlePrintln(err error, msg string) {
	if err != nil {
		log.Println(msg, err)
	}
}

// ErrHandleFatalln log error and quit
func ErrHandleFatalln(err error, msg string) {
	if err != nil {
		log.Fatalln(msg, err)
	}
}

// Executor execute input string
func Execute(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return errors.New("you need to pass the something arguments")
	} else if s == "quit" || s == "exit" {
		log.Println("Bye!")
		os.Exit(0)
	}

	cmd := exec.Command("bash", "-c", s)
	log.Println("执行命令：", "bash", "-c", s)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func ExecuteAndPrintImmediately(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return errors.New("you need to pass the something arguments")
	} else if s == "quit" || s == "exit" {
		log.Println("Bye!")
		os.Exit(0)
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd := exec.Command("bash", "-c", s)
	log.Println("执行命令：", "bash", "-c", s)
	stdoutIn, _ := cmd.StdoutPipe()
	stderrIn, _ := cmd.StderrPipe()
	var errStdout, errStderr error
	stdout := io.MultiWriter(os.Stdout, &stdoutBuf)
	stderr := io.MultiWriter(os.Stderr, &stderrBuf)
	err := cmd.Start()
	if err != nil {
		return err
	}
	go func() {
		_, errStdout = io.Copy(stdout, stdoutIn)
	}()
	go func() {
		_, errStderr = io.Copy(stderr, stderrIn)
	}()
	err = cmd.Wait()
	if err != nil {
		return err
	}
	if errStdout != nil {
		return errStdout
	}
	if errStderr != nil {
		return errStderr
	}
	return nil
}

// ExecuteAndGetResult execute input string and return the echo
func ExecuteAndGetResult(s string) (string, string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", "", errors.New("you need to pass the something arguments")
	}

	var stdout, stderr bytes.Buffer
	cmd := exec.Command("bash", "-c", s)
	log.Println("执行命令：", "bash", "-c", s)
	cmd.Stdin = os.Stdin
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", "", err
	}
	log.Println("执行结果：", strings.TrimSpace(string(stdout.Bytes())))
	return strings.TrimSpace(string(stdout.Bytes())), strings.TrimSpace(string(stderr.Bytes())), nil
}

// ExecuteAndGetResult execute input string and return the echo
func ExecuteAndGetResultCombineError(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", errors.New("you need to pass the something arguments")
	}

	var stdout, stderr bytes.Buffer
	cmd := exec.Command("bash", "-c", s)
	log.Println("执行命令：", "bash", "-c", s)
	cmd.Stdin = os.Stdin
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", err
	}
	log.Println("执行结果：", strings.TrimSpace(string(stdout.Bytes())))
	stderrStr := strings.TrimSpace(string(stderr.Bytes()))
	if stderrStr != "" {
		return "", errors.New(stderrStr)
	}
	return strings.TrimSpace(string(stdout.Bytes())), nil
}

// IsFileExist checks is the file exist or not
func IsFileExist(filepath string) (bool, error) {
	_, err := os.Stat(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
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
