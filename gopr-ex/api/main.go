package main

import (
	"advanced-blog-management-system/internal/app"
	"log"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
