package main

import "github.com/rikchilvers/gledger/cmd"

// TODO: make slices not take pointers (https://philpearl.github.io/post/bad_go_slice_of_pointers/)

func main() {
	cmd.Execute()
}
