package main

import (
	"log"
	"os"
	"path/filepath"
)

func main() {
	homedir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	config_path := filepath.Join(homedir, ".config", "forecash", "account.json")
	account := newAccount(config_path)

	tui := newTui(&account)
	tui.run()
}
