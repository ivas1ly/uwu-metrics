package flags

import (
	"flag"
	"fmt"
)

func IsFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			fmt.Printf("found flag %q, flag value %q\n", f.Name, f.Value)
			found = true
		}
	})
	return found
}
