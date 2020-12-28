package testutils

import "github.com/sirupsen/logrus"

func SetLoggerToPanicOnFatal(logger *logrus.Logger) {
	logger.ExitFunc = func(int) {
		panic("test")
	}
}
