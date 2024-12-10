# `lx`: List eXecutables

`lx` recursively searches for executables in your current directory:

```
$ lx
PATH               DESCRIPTION
scripts/script.sh  Here's some info about this script.
```

You can document utf8-encoded executables with the sigil `lx:`.
`lx` will find lines containing the sigil and extract everything following it.
This means you can write messages to `lx` in comments:

```sh
#!/usr/bin/env sh

# The following lines will be extracted by `lx`:
# lx: Here's some info about this script.
# lx: Here's some more.

echo "from scripts/script.sh"
```

The first line containing `lx:` will be used for the short description when listing executables.
The remaining lines can be shown by passing the script name as the first argument:

```
$ lx scripts/script.sh
scripts/script.sh:

Here's some info about this script.
Here's some more.
```

## Usage

```
$ lx -h
Usage of lx:
  -root string
    	Root path to start search. (default ".")
  -skip-dirs string
    	Directories to skip during search. (default ".git,node_modules")
  -timeout duration
    	Command timeout. (default 1s)
```

## Installation

```sh
go install github.com/broothie/lx@latest 
```
