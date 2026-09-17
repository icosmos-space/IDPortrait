//go:build !windows

package printer

import (
	"fmt"

	"IDPortrait/core"
)

func ListPrinters() ([]core.PrinterInfo, error) {
	return nil, fmt.Errorf("当前系统不支持枚举打印机")
}

func PrintLayout(opt core.PrintOptions) (*core.PrintResult, error) {
	return nil, fmt.Errorf("当前系统不支持本机打印")
}
