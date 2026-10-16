# stdtype

Type fast. Type pretty. Type like a monkey. Right in your terminal

`stdtype` is a [monkeytype](https://github.com/monkeytypegame/monkeytype)-like typing test that live in your shell, built on the [Charm](https://charm.sh) stack.

## Running from source

You need Go 1.27 or newer.

```sh
go run ./cmd/stdtype

go run ./cmd/stdtype --mode time --duration 60
go run ./cmd/stdtype --mode word --duration 50
go run ./cmd/stdtype --mode word --duration 50 --punctuation --numbers
go run ./cmd/stdtype --list english_1k --theme default --line 4
```

## Commands

```sh
task install-dev
task test
task lint
task --list
```

## Where things live

Your settings and settings are stored in your user config directory (`~/.config/stdtype` on macOS and Linux).
