# Better tree (arguably)

Manipulate file tree in terminal.

<img alt="Preview" src="assets/preview-w-hl.gif" width="600" />

## Installation

`bt` is currently only supported on linux and macOS. Windows support is on todo list.

```bash
go install github.com/LeperGnome/bt/cmd/bt@v1.2.1
```

Or from source

```bash
make install
```

Or download prebuilt binaries from the [latest release](https://github.com/LeperGnome/bt/releases).

## Usage

```bash
bt [flags] [directory]

Usage of bt:
      --file_preview       Enable file previews (default true)
      --highlight_indent   Highlight current indent (default true)
  -i, --in_place_render    In-place render (without alternate screen)
  -p, --padding uint       Edge padding for top and bottom (default 5)
```

Key bindings:

| key           | desc                                                           |
| ------------- | -------------------------------------------------------------- |
| j / arr down  | Select next child                                              |
| k / arr up    | Select previous child                                          |
| h / arr left  | Move up a dir                                                  |
| l / arr right | Enter selected directory                                       |
| tab           | Mark selected child and move down                              |
| shift+tab     | Mark selected child and move up                                |
| d             | Move marked children (then 'p' to paste)                       |
| y             | Copy marked children (then 'p' to paste)                       |
| D             | Delete marked child                                            |
| if / id       | Create file (if) / directory (id) in current directory         |
| r             | Rename selected child                                          |
| e             | Edit selected file in $EDITOR                                  |
| gg            | Go to top most child in current directory                      |
| G             | Go to last child in current directory                          |
| H             | Toggle hidden files in current directory                       |
| enter         | Open / close selected directory or open file (xdg-open / open) |
| esc           | Clear error message / stop current operation / drop marks      |
| ?             | Toggle help                                                    |
| q / ctrl+c    | Exit                                                           |

## Configuration

You can configure `bt` via configuration file at `$HOME/.config/bt/conf.yaml`

```yaml

padding: 5
file_preview: true
highlight_indent: true
in_place_render: false

```

## Motivation

I find myself disliking a majority of column-based terminal file managers.
The reason for that is - when I need to copy/move some files across nested subdirectories,
I constantly lose context of where I am currently, because columns always go left and right.
Even though those file managers are really mature and loaded with features (e.g. [ranger](https://github.com/ranger/ranger), [lf](https://github.com/gokcehan/lf), [xplr](https://github.com/sayanarijit/xplr), [nnn](https://github.com/jarun/nnn)), it's uneasy for me to perform simple tasks.

I like how [broot](https://github.com/Canop/broot) renders the ui, but I guess that it's mainly usable for exploring a file tree, but not manipulating it (at least I found it this way, when I had to type a target directory for `move`).

That's why I'm writing my own simple tool for simple use cases. It's currently lacking a bunch of features (see todo list below), but the fundamentals are here.

## TODO

```
Functional:
- [x] Tree rendering
- [x] File preview
- [x] Scrolling trees, that don't fit the screen
- [x] Move files
- [x] Jump into empty directories
- [x] Copy / paste files
- [x] Not reading whole file contents, only fix size
- [x] Remove files
- [~] Resolve filename conflicts (kinda done)
- [x] Sorting
- [x] "G" to go bottom and "gg" to go top
- [x] Creating files and directories
- [x] Renaming files and directories
- [x] Handle fs updates
- [~] Error handling (permissions denied, etc) (kinda)
- [x] File permission in status bar
- [x] Stylesheets
- [x] Edit selected file in editor of choice
- [x] Help
- [x] Toggle hidden directories
- [x] Image preview (half-block only)
- [x] xdg-open files
- [x] Async preview
- [x] Mark multiple files
- [ ] Image preview TGP
- [ ] Custom delete cmd
- [ ] Search
- [ ] Marked to stdout on exit
- [ ] Jump to current directory
- [ ] Go higher then local root
- [ ] Make current directory a local root
- [ ] Windows support

Fixes:
- [x] Fix size notation
- [x] Check existing name on rename
- [x] "gg" drops previous operation
- [ ] File preview ignore control chars

Maintenance:
- [x] Tests (at least a little bit)
- [x] Release artifacts
- [x] CI
- [ ] Distribution
- [ ] CONTRIBUTING.md
```
