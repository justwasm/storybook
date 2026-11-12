# storybook

A universal, pluggable component storybook and substrate for [Bubble Tea v2](https://github.com/charmbracelet/bubbletea) terminal applications.

## Overview

`storybook` decouples terminal UI components from application engines and LLM backends. Components are written as self-contained plugins that auto-register via Go's idiomatic blank import pattern:

```go
import (
    "github.com/justwasm/storybook"

    _ "your-module/plugins/basics/simple"
    _ "your-module/plugins/composer/prompt"
)

func main() {
    storybook.Run()
}
```

## Features

- **Dual-Axis Matrix**:
  - **X-Axis (State Variations)**: Fast switching between component states with number keys `1..9`.
  - **Y-Axis (Inspection Tabs)**: Seamlessly toggle between Live Demo (`d`), Source Code (`s`), and Architecture Info (`i`), or cycle with `v`.
- **Zero-Invasive Porting (`RegisterModel`)**: Wrap standard `tea.Model` with zero changes to business logic.
- **Self-Registering Plugins**: High cohesion, low coupling. Add or remove stories without modifying central code.
- **Mock Quit Interceptor**: Emulates `tea.Quit` for embedded programs, displaying a non-fatal exit banner and enabling instant restart (`r` / `space`).
- **Affine Mouse Translation**: Automatic screen-to-subcanvas coordinate mapping.
- **Multi-Theme Support**: Instant switching between Tokyo Night, Solarized Dark, Solarized Light, and Terminal 16-color palettes (`t`).
- **Search & Filtering**: Real-time component discovery with `/`.

## Porting Guide & Documentation

See [PORTING_GUIDE.md](PORTING_GUIDE.md) for a comprehensive architectural deep-dive, the 5 impedance mismatches, and copy-pasteable recipes for responsive layouts, form inputs, animations, and async I/O.

## License

MIT
