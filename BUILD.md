# termos

## Build a snapshot

To build the solution with dirty repository, use the following command with `--snapshot` parameter.

```bash
goreleaser build --clean --snapshot
```

## Typical release workflow

```bash
git add --update
```

```bash
git commit -m "fix: Change."
```

```bash
git tag "$(svu next --always)"
git push --tags
goreleaser release --clean
```

## Cookiecutter initiation

```bash
cookiecutter \
  ssh://git@github.com/lukasz-lobocki/go-cookiecutter.git \
  package_name="termos"
```

### was run with following variables

- package_name: **`termos`**;
package_short_description: `Screenshots a terminal output.`

- package_version: `0.1.0`

- author_name: `Lukasz Lobocki`;
open_source_license: `CC0 v1.0 Universal`

- __package_slug: `termos`

### on

`2025-08-18 14:39:45 +0200`
