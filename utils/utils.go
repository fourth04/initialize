// Package utils provides ...
package utils

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

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

func ErrHandlePrintln(err error, msg string) {
	if err != nil {
		log.Println(msg, err)
	}
}

func ErrHandleFatalln(err error, msg string) {
	if err != nil {
		log.Fatalln(msg, err)
	}
}

func Executor(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	} else if s == "quit" || s == "exit" {
		fmt.Println("Bye!")
		os.Exit(0)
	}

	cmd := exec.Command("bash", "-c", s)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func ExecuteAndGetResult(s string) (string, string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", "", errors.New("you need to pass the something arguments")
	}

	var stdout, stderr bytes.Buffer
	cmd := exec.Command("bash", "-c", s)
	cmd.Stdin = os.Stdin
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", "", err
	}
	return string(stdout.Bytes()), string(stderr.Bytes()), nil
}
