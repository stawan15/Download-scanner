# Download Inbox Cleaner

A cautious, open-source macOS TUI and CLI for reviewing, sorting, and safely organizing Downloads. It never deletes files automatically and requires explicit confirmation before moving anything.

## Quick start

Launch the interactive terminal interface:

```sh
cleaner tui
```

The full-screen workspace has a sidebar for Scan, Files, Duplicates, Security,
Organization, and Undo. Use `j`/`k` (or arrow keys) to navigate, `←`/`→` to
scroll long lists, `enter` to refresh a panel, `d` to select a folder, and `q`
to quit. Organizing files requires typing `MOVE`; undo requires typing `UNDO`.

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
cleaner tui
```

`cleaner organize apply` remains accepted as a legacy spelling for `--apply`, but it still needs `--yes` before files move.

`scan`, `duplicates`, and the built-in `security` check accept `--format text|json`; text is the default. ClamAV must provide `clamscan` on `PATH`.

## Safety and limits

- Only direct children of the target directory are inspected or moved.
- `organize` ignores hidden files and refuses likely source-project directories.
- Existing destination files are never overwritten.
- `trash` moves files to macOS `~/.Trash`; it does not permanently delete them.
- `undo` reverses only the latest successful organization, recorded at `~/.local/share/download-cleaner/last-organization.json`.

## Organization categories

| Destination | Examples |
| --- | --- |
| `Documents/PDF` | PDF files |
| `Documents/Office` | Word, Excel, PowerPoint, OpenDocument, RTF |
| `Documents/Text` | TXT, CSV, TSV, logs |
| `Images` | PNG, JPG, HEIC, WebP, TIFF |
| `Media/Video` | MP4, MOV, MKV, AVI, WebM |
| `Media/Audio` | MP3, M4A, WAV, FLAC, OGG |
| `Archives` | ZIP, TAR, GZ, RAR, 7Z |
| `Installers` | DMG, PKG, ISO, MSI, APK |
| `Fonts` | TTF, OTF, WOFF |
| `Design` | SVG, Figma, Photoshop, Sketch, Excalidraw |
| `Code` | Go, C#, HTML, JavaScript, JSON, YAML, Markdown and more |
| `Other` | Every unmatched direct file |

Directories and hidden files remain in place. Existing destination files are
never overwritten.

## Development

```sh
go test ./...
go vet ./...
go build ./...
```

## Releases

Pushing a version tag triggers GoReleaser, which builds the Apple Silicon and
Intel macOS archives, publishes the GitHub release, and updates the Homebrew
formula in `stawan15/homebrew-tap` with the new URL and SHA256.

The source repository needs an Actions secret named `HOMEBREW_TAP_GITHUB_TOKEN`: a
fine-grained GitHub token with **Contents: Read and write** access to the
`stawan15/homebrew-tap` repository. Create a release with:

```sh
git tag v1.2.0
git push origin v1.2.0
```

## Contributing

Contributions are welcome. Please read [CONTRIBUTING.md](CONTRIBUTING.md).

## License

This project is available under the [MIT License](LICENSE).
