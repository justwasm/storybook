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

- **Self-Registering Plugins**: High cohesion, low coupling. Add or remove stories without modifying central code.
- **Mock Quit Interceptor**: Emulates `tea.Quit` for embedded programs, displaying a non-fatal exit banner and enabling instant restart (`r` / `space`).
- **3-Pane Inspection**:
  - `[1] Live Demo`: Real-time interactive playground.
  - `[2] Source Code`: Formatted code viewer with line numbers, scrolling, and mouse drag-to-copy.
  - `[3] Architecture & Info`: Metadata, upstream references, Atomic Design levels, and state preset documentation.
- **Multi-Theme Support**: Instant switching between Tokyo Night, Solarized Dark, Solarized Light, and Terminal 16-color palettes.
- **Search & Filtering**: Real-time component discovery with `/`.

## License

MIT
