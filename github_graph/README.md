# GitHub Graph

Displays the recent GitHub contribution calendar as green squares on the front display of a Busy Bar.

The graph is rendered as one XPM bitmap with GitHub's dark contribution colors. It shows the latest 36 week columns, the maximum that fits the 72x16 front display with real 2x2 day squares. The graph is 72x14. GitHub data is cached locally and refreshed at most once per calendar day. The bitmap is vertically centered at the Busy Bar's 72x48 front display. Rendering uses the official [Go Busy Bar SDK](https://github.com/lxdb/busylib-go).

## Usage

```sh
cd github_graph
cp .env.example .env
# Set GITHUB_USERNAME and BUSY_BAR_API_TOKEN in .env
go run .
```

Settings are read exclusively from `.env`; shell environment variables are not used. `.env` is ignored by Git. `BUSY_BAR_URL` defaults to `http://10.0.4.20` when omitted.

## Settings

- `GITHUB_USERNAME` required GitHub username
- `BUSY_BAR_URL` Busy Bar API URL, default `http://10.0.4.20`
- `BUSY_BAR_API_TOKEN` optional `X-API-Token` value
- `BUSY_BAR_X` bitmap center X coordinate on the front display, default `36`
- `BUSY_BAR_Y` bitmap center Y coordinate on the front display, default `8`
- `BUSY_BAR_PRIORITY` draw priority, default `50`

Use `go run . --dry-run` to print the contribution graph to the terminal with GitHub's contribution colors.
