package main

import (
	"fmt"
)

func ToBulkString(str string) string {
	return fmt.Sprintf("$%d\r\n%s\r\n", len(str), str)
}

func ToSimpleString(str string) string {
	return fmt.Sprintf("+%s\r\n", str)
}

func ToSimpleError(str string) string {
	return fmt.Sprintf("-%s\r\n", str)
}

func ToNullBulkStrings() string {
	return "$-1\r\n"
}

func ToInteger(num int) string {
	return fmt.Sprintf(":%d\r\n", num)
}
