package main

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

const (
	visibleWeeks = 36
	cellSize     = 2
	weekGap      = 0
)

type weekKey struct {
	Start time.Time
	Row   int
}

func contributionRows(days map[string]int) ([][]int, error) {
	values := make(map[weekKey]int)
	starts := make(map[time.Time]bool)
	today := calendarToday()
	for value, level := range days {
		date, err := time.Parse(dateLayout, value)
		if err != nil || date.After(today) {
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
	if len(ordered) > visibleWeeks {
		ordered = ordered[len(ordered)-visibleWeeks:]
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

func makeXPM(rows [][]int) string {
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
	return strings.Join(lines, "\n")
}

func printGraph(rows [][]int) {
	colors := [5]string{"22;27;34", "14;68;41", "0;109;50", "38;166;65", "57;211;83"}
	for _, row := range rows {
		for _, level := range row {
			fmt.Printf("\x1b[48;2;%sm  ", colors[level])
		}
		fmt.Println("\x1b[0m")
	}
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
