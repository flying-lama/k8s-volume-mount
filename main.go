package main

import (
	"fmt"
	"k8s-volume-mount/cmd"
	"k8s-volume-mount/internal"
	"os"
)

func main() {
	// Initialize configuration
	if err := internal.Initialize(); err != nil {
		fmt.Printf("Error initializing configuration: %v\n", err)
		os.Exit(1)
	}

	// Check if we have enough arguments
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Process commands
	switch os.Args[1] {
	case "mount":
		err := cmd.MountCommand(os.Args[2:])
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

	case "unmount":
	case "cleanup":
		err := cmd.CleanupCommand(os.Args[2:])
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

	case "list":
		err := cmd.ListCommand(os.Args[2:])
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

	case "forward":
		err := cmd.ForwardCommand(os.Args[2:])
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: k8s-volume-mount [command] [options]")
	fmt.Println("\nCommands:")
	fmt.Println("  mount   -pvc=NAME [-port PORT] [-type TYPE] [-namespace NAMESPACE] [-pause-on-error] [-mount-dir DIR]  Mount a volume")
	fmt.Println("  forward -pvc=NAME [-port PORT] [-type TYPE] [-namespace NAMESPACE] Forward provider server port to local machine")
	fmt.Println("  cleanup -pvc=NAME      Unmount a volume and delete associated resources")
	fmt.Println("  list                   List mounted volumes")
	fmt.Println("\nOptions:")
	fmt.Println("  -pvc         Name of the PersistentVolumeClaim")
	fmt.Println("  -port        Specific port for LocalPort Forward (default: auto-detect)")
	fmt.Println("  -provider    Mount type: webdav, sftp, nfs (default: webdav)")
	fmt.Println("  -namespace   Namespace (optional)")
	fmt.Println("  -pause-on-error  Wait for user input on error before cleanup")
	fmt.Println("  -mount-dir   Mount directory (optional, default: ~/k8s-mounts)")
}
