package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
)

var db *sql.DB
var err error

// Version holds the current application version.
//
// This can be set using build tag to set real version number.
var Version = "0.0.1-dev"

// RootCmd represents the base command when called without any subcommands
var RootCmd *cobra.Command

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func initRootCmd() {
	if RootCmd != nil {
		return
	}
	RootCmd = &cobra.Command{
		Use:   "server",
		Short: "Server",
		Long: `By default, server will start serving using the web server with no
  arguments - which can alternatively be run by running the subcommand web.`,
		RunE: runWeb,
	}
}

func main() { //go run ./cmd/server to run app.
	initRootCmd()
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbName := os.Getenv("DB_NAME")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")

	dsn := fmt.Sprintf(
		"user=%s password=%s dbname=%s host=%s port=%s sslmode=disable",
		dbUser, dbPass, dbName, dbHost, dbPort,
	)
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("[Ping] DB is unreachable: %v", err)
	}
	log.Println("✓ DB connected")
	RootCmd.Version = Version
	Execute()
}
