# Contributing

Thanks for contributing!

This repository is maintained by **Sentiens**. For general inquiries, contact `contact@sentiens.fr`.

## Quick start

- Install Go (see `go.mod` for the version).
- Run tests: `go test ./...`
- Build the CLI: `go build ./cmd/gluster-provisioner`

## Pull requests

- Keep changes focused and include a clear description.
- Prefer adding/adjusting tests when behavior changes.
- Avoid committing generated binaries.

## Development notes

This project performs disk and mount operations and the container runs `glusterd`. Many commands require elevated privileges and should be tested carefully.

## Reporting bugs

Please include:

- What you expected to happen
- What actually happened
- Logs (sanitize anything sensitive)
- OS / kernel version and container runtime

## License

By contributing, you agree that your contributions will be licensed under the MIT License (see [LICENSE](LICENSE)).
