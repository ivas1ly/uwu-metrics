package main

import "os"

func exitFunc() {
	os.Exit(1)
}

func main() {
	exitFunc()
	os.Exit(1) // want "direct call of the os.Exit function in the main function of the main package"
}
