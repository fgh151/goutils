package main

import "github.com/runetid/go-sdk/crud"

func main() {

	App, err := crud.NewCrudApplication([]string{""})

	if err != nil {
		panic(err)
	}

	App.Run()

}
