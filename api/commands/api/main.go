package main

import (
	"flag"

	"github.com/freshteapot/learnalist-api/api/acl"
	"github.com/freshteapot/learnalist-api/api/api"
)

func main() {
	database := flag.String("database", "/tmp/api.db", "The database.")
	domain := flag.String("domain", "learnalist.net", "The domain.")
	port := flag.Int("port", 80, "Port to listen on.")
	corsAllowedOrigins := flag.String("cors-allowed-origins", "", "Use , between allowed domains.")
	flag.Parse()

	api.SetDomain(*domain)

	env := api.Env{
		Port:             *port,
		DatabaseName:     *database,
		CorsAllowOrigins: *corsAllowedOrigins
	}

	api.Run(env)
}
