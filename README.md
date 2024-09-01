# Better tree (arguably)

Manipulate file tree in terminal.

![Usage screenshot](assets/bt-usage.png)

## Installation

Currently the only way to install is by building from source...

```bash
make install
```
or
```bash
go build .
mv ./bt ~/.local/bin/bt
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
| gg            | Go to top most child in current directory              |
| G             | Go to last child in current directory                  |
| enter         | Collapse / expand selected directory                   |
| esc           | Clear error message / stop current operation           |
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
- [x] Tree rendering
- [x] File preview
- [x] Scrolling trees, that don't fit the screen
- [x] Move files
- [x] Fix current on file movement to directory which is after it
- [x] Jump into empty directories
- [x] Copy / paste files
- [x] Not reading whole file contents, only fix size
- [x] Fix strange offset with empty dir
- [x] Remove files
- [~] Resolve filename conflicts (kinda done)
- [x] Sorting
- [x] "G" to go bottom and "gg" to go top
- [x] Creating files and directories
- [x] Renaming files and directories
- [x] Handle fs updates
    - [x] Unsubscribe from directories when leaving
    - [x] Subscribe on 'enter'
    - [x] Don't manually update children
- [~] Error handling (permissions denied, etc) (kinda)
- [x] File permission in status bar
- [ ] Edit selected file in editor of choice
- [ ] Jump to current directory
- [ ] Selecting multiple files
- [ ] Go higher then local root
- [ ] Make current directory a local root
- [ ] Better style

- [ ] Sorting function as a flag?
- [ ] Inline file permissions and size?
- [ ] Search?

- [~] Project structure (kinda)
- [ ] Tests
```
