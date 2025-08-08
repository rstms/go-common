package common

import (
	"fmt"
	"log"
	"path"
	"runtime"
	"strings"
)

// use a local function alias to get the source mapping right
func Fatal(err error) error {
	pc := make([]uintptr, 1)
	count := runtime.Callers(3, pc)
	if count != 0 {
		frames := runtime.CallersFrames(pc)
		frame, _ := frames.Next()
		zframe := runtime.Frame{}
		if frame != zframe {
			_, full_function := path.Split(frame.Function)
			parts := strings.Split(full_function, ".")
			function := parts[len(parts)-1]
			_, file := path.Split(frame.File)
			err := fmt.Errorf("%s:%d %s: %v", file, frame.Line, function, err)
			return err
		}
	}
	return err
}

func Fatalf(format string, args ...interface{}) error {
	pc := make([]uintptr, 1)
	count := runtime.Callers(3, pc)
	if count != 0 {
		frames := runtime.CallersFrames(pc)
		frame, _ := frames.Next()
		zframe := runtime.Frame{}
		if frame != zframe {
			_, function := path.Split(frame.Function)
			//parts := strings.Split(function, ".")
			//function := parts[len(parts)-1]
			_, file := path.Split(frame.File)
			err := fmt.Errorf("%s:%d %s: %s", file, frame.Line, function, fmt.Sprintf(format, args...))
			return err
		}
	}
	err := fmt.Errorf(format, args...)
	return err
}

func Warning(format string, args ...interface{}) {
	msg := "WARNING: " + fmt.Sprintf(format, args...)
	log.Println(msg)
}
