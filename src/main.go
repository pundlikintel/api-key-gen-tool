package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/sirupsen/logrus"
	"strconv"
	"strings"
)

func main() {
	// Run the API key generator
	fmt.Print("API key generator\n\n")
	ctx := context.Background()
	cleanUpCountPtr := flag.String("cleanup", "", "clean up database and AWS resources")
	flag.Parse()
	if *cleanUpCountPtr == "" {
		logrus.Info("Starting Creating API keys")
		Create(ctx)
	} else {
		logrus.Info("Cleaning up")
		if strings.ToLower(*cleanUpCountPtr) == "all" {
			CleanUp(ctx)
			return
		}

		if count, err := strconv.Atoi(*cleanUpCountPtr); err == nil {
			if count < 1 {
				logrus.Errorf("Invalid count %d", count)
				return
			} else {
				CleanUp(ctx, count)
				return
			}
		} else {
			logrus.Errorf("Invalid count %d", count)
		}
	}
}
