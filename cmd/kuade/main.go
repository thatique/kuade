package main

import (
	_ "github.com/thatique/kuade/app/storage/cassandra"
	_ "github.com/thatique/kuade/app/storage/memory"
	"github.com/thatique/kuade/commands"
)

func main() {
	commands.Execute()
}
