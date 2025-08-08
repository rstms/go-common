package common

import (
	"bufio"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

func Confirm(prompt string) bool {
	if ViperGetBool("force") {
		return true
	}
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("%s [y/N]: ", prompt)
		response, err := reader.ReadString('\n')
		cobra.CheckErr(err)
		response = strings.ToLower(strings.TrimSpace(response))
		if response == "y" || response == "yes" {
			msg := ViperGetString("messages.confirm_accept")
			if msg == "" {
				msg = ConfirmAcceptMessage
			}
			if msg != "" {
				fmt.Println(msg)
			}
			return true
		} else if response == "n" || response == "no" || response == "" {
			msg := ViperGetString("messages.confirm_reject")
			if msg == "" {
				msg = ConfirmRejectMessage
			}
			if msg != "" {
				fmt.Println(msg)
			}
			return false
		}
	}
}
