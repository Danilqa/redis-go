package resp

import (
	"fmt"
	"strings"
)

func BulkString(str string) string {
	return fmt.Sprintf("$%d\r\n%s\r\n", len(str), str)
}

func SimpleString(str string) string {
	return fmt.Sprintf("+%s\r\n", str)
}

func SimpleError(str string) string {
	return fmt.Sprintf("-%s\r\n", str)
}

func NullBulkString() string {
	return "$-1\r\n"
}

func NullArray() string {
	return "*-1\r\n"
}

func Integer(num int) string {
	return fmt.Sprintf(":%d\r\n", num)
}

func Array(values []string) string {
	return fmt.Sprintf("*%d\r\n%s", len(values), strings.Join(values, ""))
}
