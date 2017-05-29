package main

import (
	"gophass_server/route"
)

func main() {
	e := route.Routing()

	e.Start(":8080")
}
