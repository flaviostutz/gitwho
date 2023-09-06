package main

import (
	"fmt"
	"os"

	"github.com/flaviostutz/gitwho/cli"
)

func main() {

	if len(os.Args) < 2 {
		fmt.Println("Usage: gitwho [changes|changes-timeline|ownership|ownership-timeline|duplicates]")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "changes":
		cli.RunChanges(os.Args)

	case "changes-timeline":
		cli.RunChangesTimeline(os.Args)

	case "ownership":
		cli.RunOwnership(os.Args)

	case "ownership-timeline":
		cli.RunOwnershipTimeline(os.Args)

	case "duplicates":
		cli.RunDuplicates(os.Args)

	default:
		fmt.Println("Usage: gitwho [changes|changes-timeline|ownership|ownership-timeline|duplicates]")
		os.Exit(1)
	}
}
