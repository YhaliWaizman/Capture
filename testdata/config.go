package main

import "os"

func main() {
	retries := os.Getenv("MAX_RETRIES")
	_ = retries
}
