package main

import (
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"
)

const (
	dateLayout       = "2006-01-02"
	contributionPath = `(?i)<td\b[^>]*class="[^"]*ContributionCalendar-day[^"]*"[^>]*>`
	datePattern      = `\bdata-date="(\d{4}-\d{2}-\d{2})"`
	levelPattern     = `\bdata-level="([0-4])"`
)

var (
	contributionRE = regexp.MustCompile(contributionPath)
	dateRE         = regexp.MustCompile(datePattern)
	levelRE        = regexp.MustCompile(levelPattern)
)

func fetchContributions(username string) (map[string]int, error) {
	today := calendarToday()
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

func calendarToday() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
}
