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

	config_directory := filepath.Dir(*config_path)
	if err := os.MkdirAll(config_directory, 0755); err != nil {
		log.Fatalf("Error creating configuration directory: %v", err)
	}

	if _, err = os.Stat(*config_path); err != nil && os.IsNotExist(err) {
		content := []byte("{}")
		if err = os.WriteFile(*config_path, content, 0644); err != nil {
			log.Fatalf("Error creating configuration file: %v", err)
		}

		log.Printf("Created configuration file: %s", *config_path)
	}

	account := newAccount(config_path)
	tui := newTui(&account)
	tui.run()
}
