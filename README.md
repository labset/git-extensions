# git-extensions

A collection of git extensions for everyday branch management workflows.

## Extensions

### git-purge

Interactively clean up local branches that have been merged or squashed into the main branch.

```bash
git purge
```

Lists branches that are safe to delete and lets you select which ones to remove.

### git-recent

Quickly switch between recently used branches.

```bash
git recent
```

Shows your recent branches and lets you select one to checkout.

## Installation

### From source

Requires Go 1.25+

```bash
go install github.com/labset/git-extensions/cmd/git-purge@latest
go install github.com/labset/git-extensions/cmd/git-recent@latest
```

## Usage

Once installed, the commands are available as git subcommands:

```bash
git purge   # Clean up merged/squashed branches
git recent  # Switch to a recent branch
```

## License

Apache-2.0
