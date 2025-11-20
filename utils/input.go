package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func ReadLine(prompt string) string {
	fmt.Print(prompt)
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return strings.TrimSpace(scanner.Text())
	}
	return ""
}
