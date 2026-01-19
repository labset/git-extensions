# git-extensions

A collection of git extensions for everyday branch management workflows.

## Extensions

### git-purge

Interactively clean up local branches that have been merged or squashed into the main branch.

```bash
git purge
```

Lists branches that are safe to delete and lets you select which ones to remove.

### git-recent (coming soon)

Quickly switch between recently used branches.

## Installation

Requires Go 1.25+

```bash
git clone https://github.com/labset/git-extensions.git
cd git-extensions
make install
```

This installs all extensions to your `$GOPATH/bin`.

## Usage

Once installed, the commands are available as git subcommands:

```bash
git purge   # Clean up merged/squashed branches
```

## Development

```bash
make build   # Build all commands to bin/
make clean   # Remove built binaries
```

## License

Apache-2.0
