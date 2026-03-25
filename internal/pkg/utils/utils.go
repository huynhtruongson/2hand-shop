package utils

import (
	"fmt"
	"runtime/debug"

	"github.com/rs/xid"
)

func HandlePanic() {
	if r := recover(); r != nil {
		fmt.Println("stacktrace from panic: \n" + string(debug.Stack()))
	}
}

func GenerateXID() string {
	return xid.New().String()
}
