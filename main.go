package main

import (
	"github.com/theun1c/unloaded-service_choice.io/services"
)

func main() {

	un := services.NewUnloader()
	un.Start()

}
