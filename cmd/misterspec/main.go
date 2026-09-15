// Command misterspec is the module's one binary entrypoint. It contains
// no logic of its own — cli.Execute builds and runs the entire command
// tree (docs/architecture-specification.md §32).
package main

import (
	"os"

	"github.com/mottamarcio/misterspec/internal/cli"
)

func main() {
	os.Exit(cli.Execute())
}
