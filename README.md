# dotlink-convert

I keep my dotfiles symlinks described in whatever format the tool of
the day wants: a flat list of `path = target` pairs for one setup, a
pile of `ln -sf` lines for another. Every time I move the config
between machines I end up hand-translating one into the other. This
converts between them instead.

Two formats:

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

## usage

Build it:

    go build -o dotlink .

Convert a manifest to a script:

    ./dotlink to-script ~/.dotlinks > setup-links.sh

Convert a script back to a manifest:

    ./dotlink to-manifest setup-links.sh > .dotlinks

Both subcommands read from stdin when the file argument is omitted or
is `-`, so pipelines work without a temp file:

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

## status

Early. Handles the common case, with unit tests for both parsers. No
detection of duplicate link paths, no support for directory-of-links
layouts like GNU Stow uses, and no validation that a target actually
exists on disk.
