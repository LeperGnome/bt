# Better tree (arguably)

Manipulate file tree in terminal.

<img alt="Preview" src="assets/preview.gif" width="600" />

## Installation

```bash
go install github.com/LeperGnome/bt/cmd/bt@v1.0.0
```

Or from source

```bash
make install
```

## Usage

```bash
bt [flags] [directory]

Flags:
  -i    In-place render (without alternate screen)
  -pad uint
        Edge padding for top and bottom (default 5)
```

Key bindings:

| key           | desc                                                   |
|---------------|--------------------------------------------------------|
| j / arr down  | Select next child                                      |
| k / arr up    | Select previous child                                  |
| h / arr left  | Move up a dir                                          |
| l / arr right | Enter selected directory                               |
| d             | Move selected child (then 'p' to paste)                |
| y             | Copy selected child (then 'p' to paste)                |
| D             | Delete selected child                                  |
| if / id       | Create file (if) / directory (id) in current directory |
| r             | Rename selected child                                  |
| e             | Edit selected file in $EDITOR                          |
| gg            | Go to top most child in current directory              |
| G             | Go to last child in current directory                  |
| enter         | Collapse / expand selected directory                   |
| esc           | Clear error message / stop current operation           |
| ?             | Toggle help                                            |
| q / ctrl+c    | Exit                                                   |

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
- [ ] Image preview
- [ ] Custom delete cmd
- [ ] Mark multiple files
- [ ] Search
- [ ] Marked to stdout on exit
- [ ] Jump to current directory
- [ ] Go higher then local root
- [ ] Make current directory a local root

Fixes:
- [ ] Fix size notation
- [ ] File preview ignore control chars

Maintenance:
- [ ] Release artifacts
- [ ] Tests
- [ ] CI
- [ ] Distribution
```
