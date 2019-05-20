package main

import (
	"flag"

	"github.com/freshteapot/learnalist-api/api/acl"
	"github.com/freshteapot/learnalist-api/api/api"
)

func main() {
	casbinConfig := flag.String("casbin-config", "./rbac_model.conf", "Path to the casbin config.")
	database := flag.String("database", "/tmp/api.db", "The database.")
	domain := flag.String("domain", "learnalist.net", "The domain.")
	port := flag.Int("port", 80, "Port to listen on.")
	corsAllowedOrigins := flag.String("cors-allowed-origins", "", "Use , between allowed domains.")
	flag.Parse()

	api.SetDomain(*domain)
	// Setup access to casbin, and then run Init().
	acl := acl.NewAclFromConfig(*casbinConfig, *database)
	acl.Init()
	env := api.Env{
		Port:             *port,
		DatabaseName:     *database,
		CorsAllowOrigins: *corsAllowedOrigins,
		Acl:              *acl,
	}

	api.Run(env)
}
