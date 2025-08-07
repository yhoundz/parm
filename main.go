/*
Copyright © 2025 A.W. <EMAIL ADDRESS>
*/

package main

import (
	// "parm/cmd"
	"fmt"
	"parm/internal/config"
)

func main() {
	err := config.Init()
	if err != nil {
		fmt.Println(err)
	}
	// cmd.Execute()
}
