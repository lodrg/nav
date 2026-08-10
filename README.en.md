<div align="center">

# nav

A terminal file navigator: arrow keys (or vim-style `hjkl`) to browse, Enter to open, and `ncd` to jump straight to the selected directory.

[简体中文](README.md) · English

![version](https://img.shields.io/github/v/tag/lodrg/nav?label=version&color=blue)
![platform](https://img.shields.io/badge/platform-macOS%20%7C%20Linux%20%7C%20Windows-lightgrey)
![size](https://img.shields.io/badge/size-1.6MB-brightgreen)
![downloads](https://img.shields.io/github/downloads/lodrg/nav/total?color=green)

</div>

- Single static binary, ~1.6 MB, zero runtime dependencies
- macOS / Linux / Windows
- File type picks the opener automatically (editor / default app / run in terminal)
- `Tab` cycles sorting: name / time / size
- Current path shown at the top of the popup

## Install

macOS / Linux:

```bash
curl -fsSL https://github.com/lodrg/nav/releases/latest/download/install.sh | sh
```

Windows PowerShell:

```powershell
irm https://github.com/lodrg/nav/releases/latest/download/install.ps1 | iex
```

The installer does everything: downloads the binary for your platform to `~/.local/bin` (`%LOCALAPPDATA%\bin` on Windows), adds it to PATH, writes the `ncd` function into your shell config (zsh → `~/.zshrc`, bash → `~/.bashrc`), and enables Tab completion for `nav` / `ncd`. Open a new terminal and you're done.

Upgrade: re-run the installer. Uninstall:

```bash
curl -fsSL https://github.com/lodrg/nav/releases/latest/download/install.sh -o /tmp/nav-install.sh
NAV_UNINSTALL=1 sh /tmp/nav-install.sh
```

## Usage

| Command | What it does |
|---|---|
| `nav` | Browse the current directory: open files, enter subdirectories |
| `ncd` | Browse and switch: selecting a directory cds the shell into it; selecting a file opens it by type |

Both share the same interface; they differ only in exit semantics. `nav` is for finding and opening files in the current directory (the popup stays open after opening), `ncd` is for going somewhere else (Enter selects and exits, handing the path to the shell). See [FAQ](#faq).

### Opening files

No configuration needed — the file type picks the opener:

| Type | Examples | Opened with |
|---|---|---|
| Text / code | `.py` `.md` `.json` `.txt` | `$EDITOR` (falls back vim → vi → nano) |
| Images / docs / media | `.png` `.pdf` `.mp4` `.docx` | System default app |
| Executables | scripts and programs with exec permission | Run in the terminal |
| Other | extensionless binaries | Default app |

Extensionless text files are detected by binary sniffing.

### Keys

| Key | Action |
|---|---|
| `↑` `↓` / `k` `j` | Move (wraps around) |
| `→` `l` `⏎` | Directory: enter; file: open |
| `←` `h` | Parent directory (cursor position remembered) |
| `Tab` | Sort: name / time (newest first) / size (largest first) |
| `space` | Preview text content |
| `.` | Toggle hidden files |
| `q` `Esc` `Ctrl+C` | Quit |

In `ncd` mode: `→` `l` enter a directory, `⏎` select (directory → cd, file → open), `p` / `q` print the current directory.

## Performance

Measured on Apple M1 Pro, medians (methodology in [PERF.md](PERF.md)):

| Scenario | nav (Go) | Python reference |
|---|---|---|
| Process startup | 19.7 ms | 42.2 ms |
| Cold start, normal directory | 35.9 ms | 56.4 ms |
| Cold start, 2050-item directory | 43.7 ms | 73.6 ms |

Single process, no daemon, instant key response. For comparison, ranger cold-starts in ~300–500 ms.

## Platforms

| Platform | Architectures | Opens files with |
|---|---|---|
| macOS | arm64 / amd64 | `open` |
| Linux | arm64 / amd64 | `xdg-open` |
| Windows | arm64 / amd64 | `cmd /c start` (Windows Terminal recommended) |

## FAQ

**Why two commands?**

A child process cannot change its parent shell's working directory, so "switching directories" must be done by a shell function: `ncd` runs `nav --print`, gets the selected path, then executes `cd`. `nav` is for browsing and opening; `ncd` is for jumping. Their Enter semantics differ.

**Text files won't open on Linux?**

Text goes through the editor fallback chain `$VISUAL` → `$EDITOR` → vim → vi → nano → code. If the system has no editor at all, you get a clear message — never a silent failure. Install one or set `export EDITOR=...`.

**What does `ncd` do with a file?**

Directory → `cd`; file → `nav --open` opens it by type. For scripting, use `nav --print` directly to get the path.

## Development

```bash
./build.sh          # cross-compile all six platforms
go test ./...       # unit tests
```

Layout: `main.go` entry and arg parsing, `app.go` state machine and open strategy, `fs.go` scan / type detection / sorting, `render.go` width-aware rendering, `keys.go` key reading, `term_*.go` terminal layer, `opener_*.go` cross-platform opening.

Design notes: UI goes to stderr, stdout carries only the result path (the contract `ncd` relies on); CJK/emoji width-aware truncation.
