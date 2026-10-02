package main

import (
	"fmt"
	"os"
	"path/filepath"

	"lesson6-homework/config"
)

func main() {
	dir, _ := os.MkdirTemp("", "cfg")
	path := filepath.Join(dir, "server.json")
	cfg := config.Config{ServerPort: 8080, Environment: "development", DebugMode: true, AdminPassword: "noleak"}
	if err := config.SaveConfig(path, cfg); err != nil {
		fmt.Println("save:", err)
		return
	}
	got, err := config.LoadConfig(path)
	if err != nil {
		fmt.Println("load:", err)
		return
	}
	fmt.Printf("port=%d env=%s debug=%v pass=%q\n", got.ServerPort, got.Environment, got.DebugMode, got.AdminPassword)
	if err := config.ValidateConfig(got); err != nil {
		fmt.Println("validate:", err)
		return
	}
	fmt.Println("ok")
}
