# Contributing

Thanks for helping improve Download Inbox Cleaner.

## Development

```sh
go test ./...
go vet ./...
go build ./...
```

Please keep changes focused, add or update tests for behavior changes, and
preserve the safety model: file-changing actions must remain explicit and
require confirmation.

## Pull requests

1. Create a branch from `main`.
2. Explain the problem and the behavior change in the pull request.
3. Include the test commands you ran.
