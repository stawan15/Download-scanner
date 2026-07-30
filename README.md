# Download Inbox Cleaner

A cautious macOS command-line tool for reviewing and organizing files directly in your Downloads folder. It never deletes files automatically and requires `--apply --yes` before moving anything.

## Install

```sh
go install github.com/stawan15/download-inbox-cleaner@latest
```

Release builds target Apple Silicon and Intel Macs. Once the project tap is published:

```sh
brew install stawan15/tap/download-inbox-cleaner
```

## Commands

```sh
cleaner scan --dir ~/Downloads
cleaner scan --format json
cleaner duplicates --dir ~/Downloads --format json
cleaner security --dir ~/Downloads
cleaner security av --dir ~/Downloads
cleaner organize --dir ~/Downloads
cleaner organize --apply --yes --dir ~/Downloads
cleaner trash --dir ~/Downloads old.zip unwanted.dmg
cleaner trash --apply --yes --dir ~/Downloads old.zip unwanted.dmg
cleaner undo --yes
```

`cleaner organize apply` remains accepted as a legacy spelling for `--apply`, but it still needs `--yes` before files move.

`scan`, `duplicates`, and the built-in `security` check accept `--format text|json`; text is the default. ClamAV must provide `clamscan` on `PATH`.

## Safety and limits

- Only direct children of the target directory are inspected or moved.
- `organize` ignores hidden files and refuses likely source-project directories.
- Existing destination files are never overwritten.
- `trash` moves files to macOS `~/.Trash`; it does not permanently delete them.
- `undo` reverses only the latest successful organization, recorded at `~/.local/share/download-cleaner/last-organization.json`.

Built-in organization recognizes Documents, Images, Archives, Installers, Fonts, Design, and Code files. Unrecognized files remain in place.

## Development

```sh
go test ./...
go vet ./...
go build ./...
```
