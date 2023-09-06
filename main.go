package main

import (
	"fmt"
	"os"

	cliChanges "github.com/flaviostutz/gitwho/cli/changes"
	cliOwnership "github.com/flaviostutz/gitwho/cli/ownership"
)

func main() {

	if len(os.Args) < 2 {
		fmt.Println("Usage: gitwho [changes|changes-timeseries|ownership|ownership-timeseries|duplicates]")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "changes":
		cliChanges.RunChanges(os.Args)

	case "changes-timeseries":
		cliChanges.RunChangesTimeseries(os.Args)

	case "ownership":
		cliOwnership.RunOwnership(os.Args)

	case "ownership-timeseries":
		cliOwnership.RunOwnershipTimeseries(os.Args)

	case "duplicates":
		cliOwnership.RunDuplicates(os.Args)

	default:
		fmt.Println("Usage: gitwho [changes|changes-timeseries|ownership|ownership-timeseries|duplicates]")
		os.Exit(1)
	}
}
