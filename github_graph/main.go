package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

func envInt(values map[string]string, name string, fallback int) (int, error) {
	value := values[name]
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer", name)
	}
	return parsed, nil
}

func run() error {
	values, err := godotenv.Read(".env")
	if err != nil {
		return err
	}
	dryRun := flag.Bool("dry-run", false, "print the contribution graph without drawing")
	flag.Parse()

	username := values["GITHUB_USERNAME"]
	if username == "" {
		return errors.New("GITHUB_USERNAME is required")
	}

	days, err := fetchContributions(username)
	if err != nil {
		return fmt.Errorf("could not fetch GitHub contributions: %w", err)
	}
	rows, err := contributionRows(days)
	if err != nil {
		return err
	}
	if *dryRun {
		printGraph(rows)
		return nil
	}

	priority, err := envInt(values, "BUSY_BAR_PRIORITY", 50)
	if err != nil {
		return err
	}
	x, err := envInt(values, "BUSY_BAR_X", 36)
	if err != nil {
		return err
	}
	y, err := envInt(values, "BUSY_BAR_Y", 8)
	if err != nil {
		return err
	}
	payload := makeDrawRequest(rows, x, y, priority)
	if err := draw(defaultString(values["BUSY_BAR_URL"], defaultBusyBar), values["BUSY_BAR_API_TOKEN"], payload); err != nil {
		return err
	}
	fmt.Printf("Displayed %s's GitHub contribution graph\n", username)
	return nil
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
