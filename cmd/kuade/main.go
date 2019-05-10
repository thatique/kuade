package main

import (
	_ "github.com/thatique/kuade/app/storage/cassandra"
	"github.com/thatique/kuade/commands"
)

func main() {
	commands.Execute()
}
