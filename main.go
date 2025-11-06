package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/guevarez30/dockit/pretty"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}

	command := os.Args[1]

	// Check if we have a pretty printer for this command
	switch command {
	case "ps":
		// Pretty print docker ps
		pretty.PrintContainers(os.Args[2:])
	case "images":
		// Pretty print docker images
		pretty.PrintImages(os.Args[2:])
	default:
		// Pass through to docker command for everything else
		runDockerCommand(os.Args[1:])
	}
}

func printUsage() {
	fmt.Println("Dockit - A prettier wrapper for Docker CLI")
	fmt.Println()
	fmt.Println("Usage: dockit [command] [options]")
	fmt.Println()
	fmt.Println("Pretty Commands (enhanced output):")
	fmt.Println("  ps              List containers with pretty formatting")
	fmt.Println("  images          List images with pretty formatting")
	fmt.Println()
	fmt.Println("All other commands are passed directly to Docker:")
	fmt.Println("  dockit run [...]         -> docker run [...]")
	fmt.Println("  dockit build [...]       -> docker build [...]")
	fmt.Println("  dockit exec [...]        -> docker exec [...]")
	fmt.Println("  etc.")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  dockit ps                    # Pretty container list")
	fmt.Println("  dockit ps -a                 # All containers (pretty)")
	fmt.Println("  dockit images                # Pretty image list")
	fmt.Println("  dockit run -d nginx          # Standard docker run")
}

func runDockerCommand(args []string) {
	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			os.Exit(exitError.ExitCode())
		}
		fmt.Fprintf(os.Stderr, "Error running docker command: %v\n", err)
		os.Exit(1)
	}
}
