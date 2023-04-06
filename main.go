package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

func main() {
	prohibitIndirectDepUpdate := os.Getenv("PROHIBIT_INDIRECT_DEP_UPDATE") == "true"

	// Read the original go.mod and go.sum files
	originalGoMod, err := os.ReadFile("go.mod")
	if err != nil {
		log.Fatalf("Failed to read go.mod: %v", err)
	}
	originalGoSum, err := os.ReadFile("go.sum")
	if err != nil {
		log.Fatalf("Failed to read go.sum: %v", err)
	}

	if prohibitIndirectDepUpdate {
		// Remove indirect blocks (except replace blocks)
		lines := strings.Split(string(originalGoMod), "\n")
		var cleanedGoMod []string
		inIndirectBlock := false
		for _, line := range lines {
			if strings.HasPrefix(line, "require (") {
				inIndirectBlock = true
				continue
			}
			if inIndirectBlock && strings.HasPrefix(line, ")") {
				inIndirectBlock = false
				continue
			}
			if !inIndirectBlock {
				cleanedGoMod = append(cleanedGoMod, line)
			}
		}
		err = os.WriteFile("go.mod", []byte(strings.Join(cleanedGoMod, "\n")), 0644)
		if err != nil {
			log.Fatalf("Failed to write cleaned go.mod: %v", err)
		}
	}

	// Run go mod tidy
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		log.Fatalf("Failed to run go mod tidy: %v", err)
	}

	// Read
	updatedGoMod, err := os.ReadFile("go.mod")
	if err != nil {
		log.Fatalf("Failed to read updated go.mod: %v", err)
	}
	updatedGoSum, err := os.ReadFile("go.sum")
	if err != nil {
		log.Fatalf("Failed to read updated go.sum: %v", err)
	}
	// Compare the original and updated files
	if string(originalGoMod) != string(updatedGoMod) || string(originalGoSum) != string(updatedGoSum) {
		fmt.Println("go.mod or go.sum files have changed after running go mod tidy. Please commit the changes.")
		os.Exit(1)
	}

	fmt.Println("Go mod check action completed successfully.")
}
