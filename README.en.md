<div align="center">

# 📂 nav

**A file navigator for your terminal** — browse with arrow keys, open with Enter, jump with `ncd`

[English](README.en.md) · [简体中文](README.md)

![version](https://img.shields.io/github/v/tag/lodrg/nav?label=version&color=blue)
![platform](https://img.shields.io/badge/platform-macOS%20%7C%20Linux%20%7C%20Windows-lightgrey)
![size](https://img.shields.io/badge/size-1.6MB-brightgreen)
![downloads](https://img.shields.io/github/downloads/lodrg/nav/total?color=green)

</div>

---

## ✨ Highlights

| | |
|---|---|
| ⚡ **Instant startup** | Single static binary, ~4 ms launch, zero runtime dependencies |
| ⌨️ **Keyboard-only** | Arrow keys + vim-style `hjkl`, hands never leave the keyboard |
| 📂 **Path always visible** | Current directory & item count shown at the top |
| 🗂️ **One-key sorting** | `Tab` cycles: name → time → size |
| 🧭 **Select & go** | `ncd` jumps your shell to the selected directory |
| 🖼️ **Smart opening** | File type picks the right handler: editor / default app / terminal |

## 🚀 Quick start

```bash
# macOS / Linux (one command, fully self-configuring)
curl -fsSL https://github.com/lodrg/nav/releases/latest/download/install.sh | sh

# Windows PowerShell
irm https://github.com/lodrg/nav/releases/latest/download/install.ps1 | iex
```

Then **open a new terminal** — you get two commands:

| Command | What it does | When to use |
|---|---|---|
| `nav` | Pop up a file list, browse & open files | Finding files **in the current** directory |
| `ncd` | Same, but the selected directory becomes your shell's cwd | Going **somewhere else** |

> The installer does everything: downloads the right binary for your platform to `~/.local/bin` → adds it to **PATH** → writes the `ncd` function into your shell config (zsh → `~/.zshrc`, bash → `~/.bashrc`, PowerShell → `$PROFILE`) → enables Tab completion for `nav`/`ncd`.

## 🖥️ What it looks like

```
$ ncd            ← type this anywhere

📂 ~/projects/dev/nav-go   (14 items · name)     ← top: path / count / sort mode
────────────────────────────────────────────
📁 dist        —      —       2026-08-08 21:32
📄 app.go      —    7.3 KB    2026-08-08 22:10   ◄ highlighted = selection
📄 README.md   —    5.0 KB    2026-08-08 22:15
────────────────────────────────────────────
↑↓/jk move · →/l ⏎ open · ←/h up · Tab sort · space preview · q quit

$              ← press q, terminal left clean
```

## 📖 Usage

### Opening files: the right handler, automatically

nav picks the most suitable way to open based on file type — no configuration needed:

| File type | Examples | Opened with |
|---|---|---|
| Text / code | `.py` `.md` `.json` `.txt` `.go` `.yml` … | **Editor** (`$EDITOR`, falls back vim → vi → nano) |
| Images / docs / media | `.png` `.jpg` `.pdf` `.mp4` `.docx` `.zip` … | **System default app** |
| Executables | scripts/bins with exec permission (`./run.sh`) | **Run in terminal**, any key to return |
| Unknown | extensionless binaries | Default app as fallback |

Fully automatic — `→` / `l` / `Enter` open anything. Want a peek first? `space` previews the text content; press again to go back.

### Key reference

| Key | Action |
|---|---|
| `↑` `↓` / `k` `j` | Move selection (wraps around) |
| `→` `l` / `Enter` | Directory → enter; file → open with auto handler |
| `←` `h` | Go to parent (**remembers your cursor position**) |
| `Tab` | Cycle sort: name → time (newest first) → size (largest first) |
| `space` | Preview text content |
| `.` | Toggle hidden files |
| `q` `Esc` `Ctrl+C` | Quit |

### `ncd` mode differences

`→` / `l` = drill into directory; `Enter` = **smart action** (directory → shell cd there; file → open by type); `p` / `q` = stay and print the current directory path.

> 💡 Why two commands? `cd` can only be executed by the shell itself (a child process can never change its parent's working directory), so "jump there" has to be wrapped in a shell function. See [FAQ](#-faq).

## 🤔 FAQ

**Why are there two commands, `nav` and `ncd`?**

It's a hard constraint of the Unix process model: `nav` is a separate process, and any `cd` it performs only affects itself — **it cannot change the parent shell's working directory**. So "select a directory and move there" must be done by a shell function:

```
ncd  →  nav --print (navigate, print the selected path)  →  ncd receives it  →  builtin cd
```

`nav` is for *seeing and opening*; `ncd` is for *going*. A physical limitation of the process model, not a design whim.

**Why does `ncd` open a file instead of printing its path?**

Since v0.2.8, `ncd` dispatches smartly: directory → `cd`; file → `nav --open` opens it by type (text→editor, executable→terminal, other→default app). For scripting, `nav --print` still prints the raw path.

**Text files won't open on Linux?**

Text goes through the editor fallback chain `$VISUAL` → `$EDITOR` → vim → vi → nano → code. If the server has no editor at all, you get a clear message — never a silent failure. Install `nano` or set `export EDITOR=your-editor`.

## ⚡ Performance

Single static binary, zero dependencies, instant startup (measured on Apple M1 Pro):

| Scenario (median) | nav (Go) | Python reference |
|---|---|---|
| Process startup | **19.7 ms** | 42.2 ms |
| Full cold start · normal dir | **35.9 ms** | 56.4 ms |
| Full cold start · 2050-item dir | **43.7 ms** | 73.6 ms |
| Binary size | **1.61 MB** | script + interpreter |

An order of magnitude faster than ranger (300–500 ms); same league as lf / nnn (compiled, lightweight).
Full report (methodology & breakdown): [PERF.md](PERF.md)

## 🔧 Install / upgrade / uninstall

```bash
# Upgrade: re-run the installer (overwrites binary + refreshes ncd function)
curl -fsSL https://github.com/lodrg/nav/releases/latest/download/install.sh | sh

# Uninstall: remove binary + ncd function + PATH entries
curl -fsSL https://github.com/lodrg/nav/releases/latest/download/install.sh -o /tmp/nav-install.sh
NAV_UNINSTALL=1 sh /tmp/nav-install.sh
```

Tunables: `NAV_URL=<prefix>` custom download source · `NAV_DEST=<dir>` custom install dir · `NAV_NO_NCD=1` skip shell config.

## 🗺️ Platform support

| Platform | Status | Opens with |
|---|---|---|
| macOS (arm64/amd64) | ✅ Full support | `open` |
| Linux (arm64/amd64) | ✅ Full support | `xdg-open` |
| Windows (arm64/amd64) | ✅ Supported (Windows Terminal recommended) | `cmd /c start` |

## 🛠️ For developers

```bash
./build.sh          # cross-compile all six platforms (needs Go ≥ 1.22)
go test ./...       # unit tests
```

```
main.go            arg parsing + TTY guard + entry
app.go             App state machine + main loop + open strategy (suspend/restore terminal)
fs.go              scan / type detection / sorting (name·time·size)
render.go          width-aware truncation + renderRegion pure func + formatRow
keys.go            key reading (ESC sequences + 100 ms timeout)
term_*.go          terminal layer (cbreak: darwin/linux/windows)
opener_*.go        default app open (open / xdg-open / cmd start)
```

**Design principles**: UI goes to stderr, stdout carries only the result path (the contract `ncd`'s command substitution relies on); cbreak not raw (Ctrl+C stays sane); CJK/emoji width-aware truncation never tears Chinese filenames; single goroutine reads keys serially.

---

<div align="center">

⭐ Like it? Star it. Love it? Tell a friend who still types `cd` the long way.

</div>
