package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/x-cray/logrus-prefixed-formatter"
	"strings"
	"testing"
)

func TestLogSaveFormater(t *testing.T) {
	tests := map[string]string{
		"MODULE": "TEST1",
		"level":  "info",
		"msg":    "xxx",
		"prefix": "TEST1",
	}
	textFormat := &prefixed.TextFormatter{
		FullTimestamp:   true,
		ForceColors:     true,
		ForceFormatting: true,
	}
	logger := EnsureLogger("TEST1")
	buf := bytes.NewBuffer([]byte{})
	mHook := NewMyHook(buf, &logrus.JSONFormatter{}, textFormat)
	logrus.AddHook(mHook)
	for key, _ := range loggers {
		mHook.moduleLevel[key] = logrus.InfoLevel
	}

	logger.Info(tests["msg"])
	saveKey := map[string]string{}
	err := json.Unmarshal(buf.Bytes(), &saveKey)
	if err != nil {
		t.Error(err)
	}
	for key, value := range tests {
		if value != saveKey[key] {
			t.Errorf("expected %s but got %s", value, saveKey[key])
		}
	}
}

func TestLogPrintFormater(t *testing.T) {
	tests := map[string]string{
		"MODULE": "TEST1",
		"level":  "INFO",
		"msg":    "xxx",
		"prefix": "TEST1",
	}
	buf := bytes.NewBuffer([]byte{})
	textFormat := &prefixed.TextFormatter{
		FullTimestamp:   true,
		ForceColors:     false,
		ForceFormatting: true,
	}
	logrus.SetFormatter(&NullFormat{})
	logrus.SetOutput(buf)
	logger := EnsureLogger("TEST1")
	mHook := NewMyHook(bytes.NewBuffer([]byte{}), &logrus.JSONFormatter{}, textFormat)
	logrus.AddHook(mHook)
	for key, _ := range loggers {
		mHook.moduleLevel[key] = logrus.InfoLevel
	}

	logger.Info(tests["msg"])
	outString := buf.String()
	fmt.Println(outString)
	for key, value := range tests {
		if !strings.Contains(outString, tests[key]) {
			t.Errorf("expected %s", value)
		}
	}
}

func TestLogLevel(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	textFormat := &prefixed.TextFormatter{
		FullTimestamp:   true,
		ForceColors:     false,
		ForceFormatting: true,
	}
	logrus.SetFormatter(&NullFormat{})
	logrus.SetOutput(buf)
	logger := EnsureLogger("TEST1")
	mHook := NewMyHook(bytes.NewBuffer([]byte{}), &logrus.JSONFormatter{}, textFormat)
	logrus.AddHook(mHook)
	for key, _ := range loggers {
		mHook.moduleLevel[key] = logrus.InfoLevel
	}

	mHook.SetLevel(logrus.FatalLevel)
	logger.Info("XXXX")
	if buf.Len() > 0 {
		t.Error("expected behavior is unable to write to the log, but written")
	}
	mHook.SetLevel(logrus.TraceLevel)
	if buf.Len() < 0 {
		t.Error("expected behavior is writed to the log, but not written")
	}
}

func TestLogModuleLevel(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	textFormat := &prefixed.TextFormatter{
		FullTimestamp:   true,
		ForceColors:     false,
		ForceFormatting: true,
	}
	logrus.SetFormatter(&NullFormat{})
	logrus.SetOutput(buf)
	logger := EnsureLogger("TEST1")
	logger2 := EnsureLogger("TEST2")
	mHook := NewMyHook(bytes.NewBuffer([]byte{}), &logrus.JSONFormatter{}, textFormat)
	logrus.AddHook(mHook)
	for key, _ := range loggers {
		mHook.moduleLevel[key] = logrus.InfoLevel
	}

	mHook.SetModulesLevel("TEST1", logrus.FatalLevel)
	logger.Info("TTTTTTT")
	logger2.Info("TTTTTTT")
	if strings.Contains(buf.String(), "TEST1") {
		t.Errorf("export module %s not log but logged", "TEST1")
	}
	if !strings.Contains(buf.String(), "TEST2") {
		t.Errorf("export module %s loged but not logged", "TEST2")
	}
}
