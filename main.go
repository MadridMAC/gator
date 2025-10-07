package main

import (
	"fmt"

	"github.com/MadridMAC/gator/internal/config"
)

func main() {
	configFile := config.Read()
	configFile.SetUser("madrid")
	updatedConfig := config.Read()
	fmt.Println(updatedConfig)
}
