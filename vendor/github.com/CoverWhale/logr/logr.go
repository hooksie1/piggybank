// Copyright 2023 Cover Whale Insurance Solutions Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logr

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

type Level struct {
	val  int
	name string
}

var (
	std           = NewLogger()
	format string = "2006-01-02T15:04:05"
)

type Logger struct {
	Level Level
	*log.Logger
	contextMessage string
}

var (
	ErrorLevel = Level{val: 0, name: "ERROR"}
	InfoLevel  = Level{val: 1, name: "INFO"}
	DebugLevel = Level{val: 2, name: "DEBUG"}
)

func getLevel(l string) Level {
	switch strings.ToLower(l) {
	case "error":
		return ErrorLevel
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	default:
		return InfoLevel
	}
}

func NewLogger() *Logger {
	level, ok := os.LookupEnv("LOG_LEVEL")
	if ok {
		level = os.Getenv("LOG_LEVEL")
	}

	l := getLevel(level)
	logger := log.Default()
	logger.SetFlags(0)
	return &Logger{
		l,
		logger,
		"",
	}
}

func (l *Logger) log(lvl Level, s interface{}) {
	if lvl.val <= l.Level.val {
		if l.contextMessage != "" {
			l.Printf(`timestamp=%s level=%s msg=%q%s`, time.Now().Format(format), lvl.name, s, l.contextMessage)
		} else {
			l.Printf(`timestamp=%s level=%s msg=%q`, time.Now().Format(format), lvl.name, s)
		}
	}
}

func (l *Logger) clone() *Logger {
	copy := *l
	return &copy
}

func (l *Logger) WithContext(s map[string]string) *Logger {
	c := l.clone()
	context := l.contextMessage
	for k, v := range s {
		context = fmt.Sprintf("%s %s=%s", context, k, v)
	}
	c.contextMessage = context

	return c

}

func (l *Logger) Error(s interface{}) {
	l.log(ErrorLevel, s)
}

func (l *Logger) Errorf(format string, s ...interface{}) {
	f := fmt.Sprintf(`%s`, format)
	m := fmt.Sprintf(f, s...)
	l.log(ErrorLevel, m)
}

func (l *Logger) Info(s interface{}) {
	l.log(InfoLevel, s)
}

func (l *Logger) Infof(format string, s ...interface{}) {
	f := fmt.Sprintf(`%s`, format)
	m := fmt.Sprintf(f, s...)
	l.log(InfoLevel, m)
}

func (l *Logger) Debug(s interface{}) {
	l.log(DebugLevel, s)
}

func (l *Logger) Debugf(format string, s ...interface{}) {
	f := fmt.Sprintf(`%s`, format)
	m := fmt.Sprintf(f, s...)
	l.log(DebugLevel, m)
}

func (l *Logger) Fatal(s interface{}) {
	m := fmt.Sprintf(`level=FATAL msg=%s`, s)
	l.Logger.Fatal(m)
}

func (l *Logger) Fatalf(format string, s ...interface{}) {
	f := fmt.Sprintf(`level=FATAL msg=%s`, format)
	m := fmt.Sprintf(f, s...)
	l.Logger.Fatal(m)
}

func Error(s interface{}) {
	std.Error(s)
}

func Errorf(format string, s ...interface{}) {
	std.Errorf(format, s...)
}

func Info(s interface{}) {
	std.Info(s)
}

func Infof(format string, s ...interface{}) {
	std.Infof(format, s...)
}

func Debug(s interface{}) {
	std.Debug(s)
}

func Debugf(format string, s ...interface{}) {
	std.Debugf(format, s...)
}

func Fatal(s interface{}) {
	std.Fatal(s)
}

func Fatalf(format string, s ...interface{}) {
	std.Fatalf(format, s...)
}

// GetCaller is a helper function to get the function name to provide context for an error
func GetCaller() string {
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		return "function name unknown"
	}

	funcName := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	return funcName[1]
}
