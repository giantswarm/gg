package colour

import "fmt"

const (
	format = "\033[38;5;%dm%s\033[39;49m"
)

func Blue(s string) string {
	return sprintf(117, s)
}

func Green(s string) string {
	return sprintf(114, s)
}

func Red(s string) string {
	return sprintf(161, s)
}

func sprintf(c int, s string) string {
	return fmt.Sprintf(format, c, s)
}
