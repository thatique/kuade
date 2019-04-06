package main

import (
	"github.com/thatique/kuade/cmd"

	_ "github.com/thatique/kuade/app/storage/mongo"
	_ "github.com/thatique/kuade/pkg/mailer/smtp"
)

func main() {
	cmd.RootCmd.Execute()
}
