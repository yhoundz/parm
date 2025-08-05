/*
Copyright © 2025 A.W. <EMAIL ADDRESS>
*/

package main

import (
	"parm/cmd"
	"parm/internal/config"
)

func main() {
	config.Init()
	cmd.Execute()
}
