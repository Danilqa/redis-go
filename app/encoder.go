package main

import (
	"fmt"
)

func DecodeString(str string) string {
	return fmt.Sprintf("$%d\r\n%s\r\n", len(str), str)
}
