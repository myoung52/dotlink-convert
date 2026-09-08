# dotlink-convert

I keep my dotfiles symlinks described in whatever format the tool of
the day wants: a flat list of `path = target` pairs for one setup, a
pile of `ln -sf` lines for another. Every time I move the config
between machines I end up hand-translating one into the other. This
converts between them instead.

Three formats:

**manifest** - a flat text file, one link per line:

```
# ~/.dotlinks
~/.vimrc = dotfiles/vim/vimrc
~/.zshrc = dotfiles/zsh/zshrc
~/.config/nvim = dotfiles/nvim
```

**script** - a POSIX shell script of `ln -sf` commands:

```sh
#!/bin/sh
ln -sf dotfiles/vim/vimrc ~/.vimrc
ln -sf dotfiles/zsh/zshrc ~/.zshrc
ln -sf dotfiles/nvim ~/.config/nvim
```

**stow** - a GNU Stow package directory on disk, read directly rather
than as a text file. Each top-level entry becomes one link into a
target directory:

```
dotfiles/
  .vimrc
  .config/
```

read with `-target ~` produces `~/.vimrc = dotfiles/.vimrc` and
`~/.config = dotfiles/.config`, the same tree-folding stow itself does
when the target is otherwise empty. This is read-only - dotlink has no
`to-stow` output, since generating a package directory means moving
real files around, not just printing text.

## usage

Build it:

    go build -o dotlink .

Convert a manifest to a script:

    ./dotlink to-script ~/.dotlinks > setup-links.sh

Convert a script back to a manifest:

    ./dotlink to-manifest setup-links.sh > .dotlinks

Check a manifest against what's actually on disk:

    ./dotlink check ~/.dotlinks

For each entry this reports whether the link path is missing, is a
regular file or directory instead of a symlink, points somewhere
other than the expected target, or is a symlink to a target that
doesn't exist. It exits non-zero if anything is wrong.

All three subcommands take `-format` to pick the input format:
`manifest` (the default, `script` for `to-manifest`), `script`, or
`stow`. With `-format stow` the file argument is a package directory
instead of a file, and `-target` sets the directory its links point
into (default: your home directory):

    ./dotlink to-manifest -format stow -target ~ dotfiles/vim > vim.dotlinks
    ./dotlink check -format stow dotfiles/vim

The manifest and script subcommands read from stdin when the file
argument is omitted or is `-`, so pipelines work without a temp file
(this doesn't apply to `-format stow`, which always needs a real
directory):

    cat .dotlinks | ./dotlink to-script
    curl -s https://example.com/setup-links.sh | ./dotlink to-manifest

## format notes

Manifest lines are `path = target`; blank lines and `#` comments are
skipped. The script reader only looks at lines starting with `ln` -
flags are ignored, and whatever two path arguments are left are read
as target and link path, in that order, matching `ln`'s own argument
order.

Paths are passed through exactly as written by default - that's
between you and whatever shell runs the output. Pass `-expand` to
have dotlink expand a leading `~` or `$HOME` itself, in both paths
and targets:

    ./dotlink to-script -expand ~/.dotlinks > setup-links.sh

That's useful when the output is going somewhere that won't do the
expansion for you, like a manifest read back in by something other
than a shell.

If two entries end up with the same link path - whether that's
written that way or only happens after `-expand` resolves `~` and
`$HOME` to the same place - both subcommands refuse to write output
and report every duplicated path instead of silently letting the
last one win.

## status

Early. Handles the common case, with unit tests for the manifest and
script parsers, `check`, and the stow reader. Stow support only
covers the clean-target case - it doesn't unfold a directory that
already exists for real on the target side, since that decision needs
to inspect live target state rather than just the package.
