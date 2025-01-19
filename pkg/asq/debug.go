package asq

import "fmt"

func IsDebugging() bool {
	return false
}

func Debug(format string, args ...interface{}) {
	if IsDebugging() {
		fmt.Printf(format+"\n", args...)
	}
}
