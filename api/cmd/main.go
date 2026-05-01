package main

import (
	"dashboard/api/internal/app"
	"dashboard/api/internal/config"
	"flag"
	"fmt"
	"log"
	"runtime"
	"runtime/debug"

	"github.com/joho/godotenv"
)

func init() {

	info, _ := debug.ReadBuildInfo()
	fmt.Println("Go version:", info.GoVersion, runtime.GOARCH)

	runtimeFlag := flag.String("runtime", "", "specify runtime flag value: -runtime=<value>")
	flag.Parse()

	if *runtimeFlag == string(config.DevRuntime) {
		if err := godotenv.Load(".env"); err != nil {
			log.Fatal(err)
		}
		log.Println("⚙️ Dev runtime, loaded .env")
	}
}

func main() {
	cfg := config.New()

	application := app.New(cfg)

	application.Run()
}
