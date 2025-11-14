package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/carped99/steampipe-plugin-openfga/openfga"
)

const defaultConfigFile = "testdata.json"

// config represents the JSON configuration structure
type config struct {
	Connection struct {
		Endpoint string `json:"endpoint"`
		StoreID  string `json:"store_id"`
	} `json:"connection"`
	Generation struct {
		Users   int `json:"users"`
		Docs    int `json:"docs"`
		Folders int `json:"folders"`
	} `json:"generation"`
	Tuples struct {
		Relations []string `json:"relations"`
	} `json:"tuples"`
	Output struct {
		ShowSamples bool `json:"show_samples"`
		SampleSize  int  `json:"sample_size"`
		Verbose     bool `json:"verbose"`
	} `json:"output"`
}

func main() {
	configFile := flag.String("config", defaultConfigFile, "Path to configuration JSON file")
	flag.Parse()

	cfg := loadConfig(*configFile)
	ctx := context.Background()

	data, err := generateTestData(cfg)
	if err != nil {
		log.Fatalf("Failed to generate test data: %v", err)
	}

	fmt.Println("\n=== Creating Tuples in OpenFGA ===")

	client, err := openfga.NewClient(ctx, openfga.Config{
		Endpoint: cfg.Connection.Endpoint,
		StoreId:  &cfg.Connection.StoreID,
	})
	if err != nil {
		log.Fatalf("Failed to create OpenFGA client: %v", err)
	}
	defer client.Close()

	if err := CreateTuplesFromTestData(ctx, client, cfg.Connection.StoreID, data, cfg.Tuples.Relations); err != nil {
		log.Fatalf("Failed to create tuples: %v", err)
	}

	fmt.Println("\n=== Completed Successfully ===")
}

func loadConfig(filename string) config {
	var configPath string

	// If filename is relative, try to find it relative to this file's location
	if !filepath.IsAbs(filename) {
		// Try current directory first
		if _, err := os.Stat(filename); err == nil {
			configPath = filename
		} else {
			// Try testdata directory
			testdataPath := filepath.Join("testdata", filename)
			if _, err := os.Stat(testdataPath); err == nil {
				configPath = testdataPath
			} else {
				configPath = filename // fallback to original
			}
		}
	} else {
		configPath = filename
	}

	absPath, err := filepath.Abs(configPath)
	if err != nil {
		log.Fatalf("Failed to resolve config path: %v", err)
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		log.Fatalf("Failed to read config file %s: %v", absPath, err)
	}

	var cfg config
	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}

	if cfg.Connection.Endpoint == "" {
		log.Fatal("Endpoint is required")
	}
	if cfg.Connection.StoreID == "" {
		log.Fatal("Store ID is required")
	}
	if cfg.Generation.Users < 0 || cfg.Generation.Docs < 0 || cfg.Generation.Folders < 0 {
		log.Fatal("Entity counts must be non-negative")
	}

	fmt.Println("=== OpenFGA Test Data Generator ===")
	fmt.Printf("\nEndpoint: %s\n", cfg.Connection.Endpoint)
	fmt.Printf("Store ID: %s\n", cfg.Connection.StoreID)
	fmt.Printf("Users: %d, Docs: %d, Folders: %d\n\n", cfg.Generation.Users, cfg.Generation.Docs, cfg.Generation.Folders)

	return cfg
}
