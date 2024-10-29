package cmd

import (
	"github.com/dhbin/ai-connect/internal/app/chatgpt"
	"github.com/spf13/cobra"
)

var (
	mirror     bool
	chatgptCmd = &cobra.Command{
		Use:   "chatgpt",
		Short: "chatgpt相关功能",
		Run: func(cmd *cobra.Command, args []string) {
			if mirror {
				chatgpt.RunMirror()
			}
		},
	}
)
