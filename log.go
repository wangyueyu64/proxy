package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
)

var logger *logrus.Logger

func LogInit() error {

	logfile, err := os.OpenFile("C:\\Users\\25781\\GolandProjects\\Server\\log.txt", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		fmt.Printf("%s\r\n", err.Error())
		os.Exit(-1)
	}

	logger = logrus.New()
	logger.SetOutput(logfile)

	return nil
}
