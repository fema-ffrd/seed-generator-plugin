package main

import (
	"log"

	"github.com/fema-ffrd/cc-go-sdk"
	tiledb "github.com/fema-ffrd/cc-go-sdk/tiledb-store"
	_ "github.com/fema-ffrd/seed-generator-plugin/internal/actions"
)

var commit string
var date string

func main() {
	//register tiledb
	cc.DataStoreTypeRegistry.Register("TILEDB", tiledb.TileDbEventStore{})

	pm, err := cc.InitPluginManager()
	if err != nil {
		log.Fatalf("Unable to initialize the plugin manager: %s\n", err)
	}

	pm.Logger.Info("Seed Generator", "version", commit, "build-date", date)

	err = pm.RunActions()
	if err != nil {
		pm.Logger.Fatalf("error running actions: %s", err.Error())
	}

}
