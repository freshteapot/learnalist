package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/freshteapot/learnalist-api/api/client"
)

func main() {
	username := flag.String("username", "", "The user you want to login with.")
	//@todo make this more secure, by at least hiding the password in a config file.
	password := flag.String("password", "", "The password for the user.")

	server := flag.String("server", "https://learnalist.net/api", "The server.")

	showSupported := flag.Bool("show-supported", false, "When set, show the api endpoints supported by the client.")
	flag.Parse()

	if *showSupported {
		supported()
		os.Exit(0)
	}

	config := client.Config{
		Server:   *server,
		Username: *username,
		Password: *password,
	}

	client := client.Client{
		Config: config,
	}

	rootResponse, _ := client.GetRoot()
	fmt.Println(rootResponse)
	versionResponse, _ := client.GetVersion()
	fmt.Println(versionResponse)
}

func supported() {
	apiSupported := `
	/
	/version
`
	fmt.Println(apiSupported)
}
