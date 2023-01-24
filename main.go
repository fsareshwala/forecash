package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
)

func main() {
	homedir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	default_config_path := filepath.Join(homedir, ".config", "forecash", "account.json")

	config_path := flag.String("config", default_config_path, "account configuration file")
	flag.Parse()

	account := newAccount(config_path)
	tui := newTui(&account)
	tui.run()
}
