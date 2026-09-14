package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	busylib "github.com/lxdb/busylib-go"
)

const (
	appName          = "github_graph"
	defaultBusyBar   = "http://10.0.4.20"
	dateLayout       = "2006-01-02"
	visibleWeeks     = 36
	cellSize         = 2
	weekGap          = 0
	contributionPath = `(?i)<td\b[^>]*class="[^"]*ContributionCalendar-day[^"]*"[^>]*>`
	datePattern      = `\bdata-date="(\d{4}-\d{2}-\d{2})"`
	levelPattern     = `\bdata-level="([0-4])"`
)

type weekKey struct {
	Start time.Time
	Row   int
}

var (
	contributionRE = regexp.MustCompile(contributionPath)
	dateRE         = regexp.MustCompile(datePattern)
	levelRE        = regexp.MustCompile(levelPattern)
)

func fetchContributions(username string) (map[string]int, error) {
	today := time.Now()
	start := today.AddDate(0, 0, -365)
	endpoint := fmt.Sprintf(
		"https://github.com/users/%s/contributions?from=%s&to=%s",
		url.PathEscape(username), start.Format(dateLayout), today.Format(dateLayout),
	)
	request, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "text/html")
	request.Header.Set("User-Agent", "busyapps-github-graph/1.0")
	client := http.Client{Timeout: 20 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("GitHub returned %s", response.Status)
	}
	page, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	contributions := make(map[string]int)
	for _, cell := range contributionRE.FindAllString(html.UnescapeString(string(page)), -1) {
		dateMatch := dateRE.FindStringSubmatch(cell)
		levelMatch := levelRE.FindStringSubmatch(cell)
		if len(dateMatch) == 2 && len(levelMatch) == 2 {
			level, _ := strconv.Atoi(levelMatch[1])
			contributions[dateMatch[1]] = level
		}
	}
	if len(contributions) == 0 {
		return nil, fmt.Errorf("no contribution days found for GitHub user %q", username)
	}
	return contributions, nil
}

func contributionWeeks(days map[string]int) ([][]int, error) {
	values := make(map[weekKey]int)
	starts := make(map[time.Time]bool)
	for value, level := range days {
		date, err := time.Parse(dateLayout, value)
		if err != nil {
			continue
		}
		start := date.AddDate(0, 0, -int(date.Weekday()))
		start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC)
		values[weekKey{Start: start, Row: int(date.Weekday())}] = max(0, min(4, level))
		starts[start] = true
	}
	ordered := make([]time.Time, 0, len(starts))
	for start := range starts {
		ordered = append(ordered, start)
	}
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Before(ordered[j]) })
	if len(ordered) > 53 {
		ordered = ordered[len(ordered)-53:]
	}
	if len(ordered) == 0 {
		return nil, errors.New("contribution data did not contain valid dates")
	}
	rows := make([][]int, 7)
	for row := range rows {
		rows[row] = make([]int, len(ordered))
		for column, start := range ordered {
			rows[row][column] = values[weekKey{Start: start, Row: row}]
		}
	}
	return rows, nil
}

func makeXPM(days map[string]int) (string, error) {
	rows, err := contributionWeeks(days)
	if err != nil {
		return "", err
	}
	if len(rows[0]) > visibleWeeks {
		for i := range rows {
			rows[i] = rows[i][len(rows[i])-visibleWeeks:]
		}
	}
	width := len(rows[0])*cellSize + (len(rows[0])-1)*weekGap
	height := len(rows) * cellSize
	lines := []string{
		"! XPM2",
		fmt.Sprintf("%d %d 6 1", width, height),
		". c none",
		"0 c #161b22",
		"1 c #0e4429",
		"2 c #006d32",
		"3 c #26a641",
		"4 c #39d353",
	}
	for _, row := range rows {
		for y := 0; y < cellSize; y++ {
			var line strings.Builder
			for column, level := range row {
				if column > 0 {
					line.WriteString(strings.Repeat(".", weekGap))
				}
				line.WriteString(strings.Repeat(string(byte('0'+level)), cellSize))
			}
			lines = append(lines, line.String())
		}
	}
	return strings.Join(lines, "\n"), nil
}

func makeDrawRequest(days map[string]int, x, y, priority int) (busylib.DisplayElements, error) {
	xpm, err := makeXPM(days)
	if err != nil {
		return busylib.DisplayElements{}, err
	}
	element := busylib.NewXPMBitmapElement("contributions", xpm)
	element.Align = busylib.DisplayAlignCenter
	element.X = &x
	element.Y = &y
	element.Display = busylib.DisplayFront
	timeout := 0
	element.Timeout = &timeout
	request := busylib.NewDisplayElements(appName, element)
	request.Priority = priority
	return request, nil
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

func envInt(name string, fallback int) (int, error) {
	value := os.Getenv(name)
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
	dryRun := flag.Bool("dry-run", false, "print the draw requests without drawing")
	flag.Parse()

	priority, err := envInt("BUSY_BAR_PRIORITY", 50)
	if err != nil {
		return err
	}

	username := os.Getenv("GITHUB_USERNAME")
	if username == "" {
		return errors.New("GITHUB_USERNAME is required")
	}

	var days map[string]int

	days, err = fetchContributions(username)
	if err != nil {
		return fmt.Errorf("could not fetch GitHub contributions: %w", err)
	}
	x, err := envInt("BUSY_BAR_X", 36)
	if err != nil {
		return err
	}
	y, err := envInt("BUSY_BAR_Y", 8)
	if err != nil {
		return err
	}
	payload, err := makeDrawRequest(days, x, y, priority)
	if err != nil {
		return err
	}
	if *dryRun {
		data, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}
	if err := draw(defaultString(os.Getenv("BUSY_BAR_URL"), defaultBusyBar), os.Getenv("BUSY_BAR_API_TOKEN"), payload); err != nil {
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
