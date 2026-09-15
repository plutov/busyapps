package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	busylib "github.com/lxdb/busylib-go"
)

const (
	appName        = "ascii_text"
	defaultBusyBar = "http://10.0.4.20"
)

func envInt(values map[string]string, name string, fallback int) (int, error) {
	if values[name] == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(values[name])
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer", name)
	}
	return value, nil
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func draw(baseURL, token string, payload busylib.DisplayElements) error {
	options := []busylib.Option{busylib.WithBaseURL(baseURL)}
	if token != "" {
		options = append(options, busylib.WithLocalAccessToken(token))
	}
	client, err := busylib.NewClient(options...)
	if err != nil {
		return err
	}
	return client.Display().Draw(context.Background(), payload)
}

func run() error {
	values, err := godotenv.Read(".env")
	if err != nil {
		return err
	}
	dryRun := flag.Bool("dry-run", false, "print the banner without drawing")
	flag.Parse()

	text := strings.ToUpper(strings.TrimSpace(values["BUSY_BAR_TEXT"]))
	if text == "" {
		return errors.New("BUSY_BAR_TEXT is required")
	}
	xpm, err := makeXPM(text)
	if err != nil {
		return err
	}
	if *dryRun {
		printBanner(xpm)
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
	element := busylib.NewXPMBitmapElement("text", xpm)
	element.Align = busylib.DisplayAlignCenter
	element.X = &x
	element.Y = &y
	element.Display = busylib.DisplayFront
	timeout := 0
	element.Timeout = &timeout
	payload := busylib.NewDisplayElements(appName, element)
	payload.Priority = priority
	if err := draw(defaultString(values["BUSY_BAR_URL"], defaultBusyBar), values["BUSY_BAR_API_TOKEN"], payload); err != nil {
		return err
	}
	fmt.Printf("Displayed %s\n", text)
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
