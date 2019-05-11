package commands

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

func secretKeyCommand() *cobra.Command {
	var gen64Bytes bool

	// secretKeyCommand
	var secretKeyCommand = &cobra.Command{
		Use:   "secret",
		Short: "secret key management",
		Long:  `Secret key management, generate secret key etc`,
		Run: func(cmd *cobra.Command, arg []string) {
			cmd.Usage()
		},
	}
	var generateCommand = &cobra.Command{
		Use:   "generate",
		Short: `generate secret key`,
		Long:  `generate key, 32 or 64 bytes length`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var len = 32
			if gen64Bytes {
				len = 64
			}
			k := make([]byte, len)
			if _, err := io.ReadFull(rand.Reader, k); err != nil {
				return err
			}
			fmt.Println("base64:" + base64.StdEncoding.EncodeToString(k))
			return nil
		},
	}

	generateCommand.Flags().BoolVarP(&gen64Bytes, "long", "L", false, "Generate the 64 bytes key version")
	secretKeyCommand.AddCommand(generateCommand)

	return secretKeyCommand
}
