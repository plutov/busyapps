# ASCII Text

Renders one uppercase word as a centered 5×7 ASCII-style banner on the Busy Bar front display.

The banner uses a dark, scanlined background with a sparse pixel texture, warm white letters, and a one-pixel pink drop shadow. Short words automatically render at double scale; longer words use the native pixel grid.

## Usage

```sh
cd ascii_text
cp .env.example .env
# Set BUSY_BAR_TEXT and connection settings
 go run .
```

Use `go run . --dry-run` to preview the colored banner in the terminal.

## Settings

- `BUSY_BAR_TEXT` required text; it is converted to uppercase. Supports `A-Z`, `0-9`, spaces, and `!`.
- `BUSY_BAR_URL` Busy Bar URL, default `http://10.0.4.20`
- `BUSY_BAR_API_TOKEN` optional `X-API-Token` value
- `BUSY_BAR_X` bitmap center X coordinate, default `36`
- `BUSY_BAR_Y` bitmap center Y coordinate, default `8`
- `BUSY_BAR_PRIORITY` draw priority, default `50`
