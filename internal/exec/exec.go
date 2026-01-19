package exec

import (
	"os/exec"
	"strings"
)

func WithOutput(command string, handler func(output string) error) error {
	args := strings.Fields(command)
	cmd := exec.Command(args[0], args[1:]...)
	out, err := cmd.Output()
	if err != nil {
		return err
	}
	return handler(strings.TrimSpace(string(out)))
}
