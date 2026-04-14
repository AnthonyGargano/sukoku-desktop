# Sudoku Desktop

A native desktop Sudoku game built with [Go](https://go.dev/) and [Fyne v2](https://fyne.io/). Featuring a fast puzzle generator with guaranteed unique solutions, pencil notes, smart hints, a game timer, leaderboards, and a polished dark-mode UI.

---

## Features

- **Uniquely-solvable puzzles** — every generated puzzle has exactly one valid solution, verified with a bitmask backtracking solver using Minimum Remaining Values (MRV) heuristic
- **Three difficulty levels** — Easy (40 clues), Medium (32), Hard (25)
- **Three themes** — Light, Medium (dark), Dark
- **Pencil notes** — toggle note mode and mark candidate digits in standard reading order (1–9)
- **Smart hints** — logical step-by-step hints (naked singles, hidden singles) with a random fallback reveal; 3 hints per puzzle
- **Error tracking** — wrong entries flash red and remain persistently highlighted; you lose after 5 mistakes (Game Over)
- **Wrong cell highlighting** — incorrect cells stay red across all selection and highlight states
- **Number palette** — shows remaining count per digit; completed digits are disabled
- **Game timer** — pauses when you pause; records your best times
- **Leaderboard** — top 5 times per difficulty saved locally, sorted by time then errors
- **Keyboard support** — `1–9` to enter digits, `Del`/`Backspace`/`0` to clear, `H` for hint, `N` for notes toggle, `P` to pause
- **Rotating log files** — daily logs auto-compressed after 7 days, deleted after 1 year

---

## Screenshots

> *(Medium theme shown)*

| Board in play | Game Over |
|---|---|
| ![Sudoku board](.github/screenshot_board.png) | ![Game Over](.github/screenshot_gameover.png) |

---

## Requirements

| Tool | Version |
|---|---|
| Go | 1.23+ |
| Fyne | v2.7.0 (fetched automatically) |

**Platform dependencies for Fyne** (needed to compile from source):

- **Windows**: No extra dependencies
- **macOS**: Xcode command-line tools (`xcode-select --install`)
- **Linux**: `gcc`, `libgl1-mesa-dev`, `xorg-dev`

See the [Fyne Getting Started](https://docs.fyne.io/started/) guide for full details.

---

## Building from Source

```bash
# Clone the repo
git clone https://github.com/AnthonyGargano/sukoku-desktop.git
cd sukoku-desktop

# Fetch dependencies
go mod download

# Build a native executable (no console window on Windows)
go build -ldflags="-H windowsgui" -o Sudoku.exe    # Windows
go build -o sudoku                                  # macOS / Linux

# Run tests
go test ./...
```

---

## Running

Simply double-click the built executable, or run it from a terminal:

```bash
./Sudoku.exe   # Windows
./sudoku       # macOS / Linux
```

The leaderboard is saved as `sudoku_scores.json` next to the executable. Logs are written to a `logs/` directory alongside it.

---

## How to Play

1. Select a **difficulty** and press **New** to generate a puzzle.
2. Click any empty cell to select it, then press a number key **1–9** (or click a palette button) to fill it in.
3. Correct entries are accepted silently. Wrong entries **flash red** and stay marked — you have **5 mistakes** before it's Game Over.
4. Use **Notes: ON** (or press `N`) to toggle pencil-mark mode. Numbers entered in note mode appear as small candidates in the cell instead of filling it.
5. Press **Hint** (or `H`) to get a logical hint. If no logical step is available, a random empty cell is revealed. You get 3 hints per puzzle.
6. Press **Pause** (or `P`) to hide the board and stop the timer.
7. Complete the board correctly to stop the clock and record your score on the leaderboard.

### Keyboard Shortcuts

| Key | Action |
|---|---|
| `1` – `9` | Enter digit (or toggle note in note mode) |
| `Del` / `Backspace` / `0` | Clear selected cell |
| `H` | Hint |
| `N` | Toggle Notes mode |
| `P` | Pause / Resume |

---

## Project Structure

```
.
├── main.go                   # Entry point; initialises logging and panic recovery
├── ui.go                     # Fyne UI, app state, game logic, highlights, timer
├── engine.go                 # Core Grid type, basic backtracking solver & generator
├── engine_opt.go             # Optimised bitmask solver (MRV) + fast puzzle generator
├── engine_accuracy_test.go   # Uniqueness and correctness tests
├── engine_bench_test.go      # Benchmark tests
├── hint.go                   # Logical hint engine (naked/hidden singles)
├── leaderboard.go            # Score persistence (JSON)
├── logging.go                # Rotating file logger with gzip compression
└── theme_custom.go           # Custom Fyne theme (Medium dark variant)
```

---

## License

This project is open source. Feel free to fork, modify, and distribute.
