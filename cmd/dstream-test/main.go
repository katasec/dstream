package main

import (
	"context"
	"fmt"
	"os"

	"github.com/katasec/dstream/internal/devinfra"
)

func main() {
	if len(os.Args) < 3 {
		usage()
		return
	}

	section := os.Args[1]
	action := os.Args[2]

	ctx := context.Background()

	switch section {
	case "mssql":
		switch action {
		case "up":
			if err := devinfra.StartMSSQL(ctx); err != nil {
				fmt.Println("ERROR:", err)
				os.Exit(1)
			}
			fmt.Println("MSSQL container started.")

		case "down":
			if err := devinfra.StopMSSQL(ctx); err != nil {
				fmt.Println("ERROR:", err)
				os.Exit(1)
			}
			fmt.Println("MSSQL container stopped.")

		default:
			fmt.Println("Unknown action:", action)
			usage()
		}

	default:
		fmt.Println("Unknown section:", section)
		usage()
	}
}

func usage() {
	fmt.Println("Usage:")
	fmt.Println("  dstream-test mssql up")
	fmt.Println("  dstream-test mssql down")
}
