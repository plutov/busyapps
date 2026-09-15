package main

import (
	"context"

	busylib "github.com/lxdb/busylib-go"
)

const (
	appName        = "github_graph"
	defaultBusyBar = "http://10.0.4.20"
)

func makeDrawRequest(rows [][]int, x, y, priority int) busylib.DisplayElements {
	xpm := makeXPM(rows)
	element := busylib.NewXPMBitmapElement("contributions", xpm)
	element.Align = busylib.DisplayAlignCenter
	element.X = &x
	element.Y = &y
	element.Display = busylib.DisplayFront
	timeout := 0
	element.Timeout = &timeout
	request := busylib.NewDisplayElements(appName, element)
	request.Priority = priority
	return request
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
