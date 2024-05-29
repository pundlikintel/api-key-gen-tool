package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/sirupsen/logrus"
)

func main() {
	// Run the API key generator
	fmt.Print("API key generator\n\n")
	ctx := context.Background()
	CleanUpPtr := flag.Bool("cleanup", false, "clean up database and AWS resources")
	flag.Parse()
	if !*CleanUpPtr {
		logrus.Info("Starting Creating API keys")
		Create(ctx)
	} else {
		logrus.Info("Cleaning up")
		CleanUp(ctx)
	}
}
