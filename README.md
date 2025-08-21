# termos ![Static](https://img.shields.io/badge/plajta-werwa-chartreuse?style=for-the-badge&labelColor=cyan)

Screenshots a terminal output.
```bash
termos shot --filename example -- ls --color=always -acl /

Flags:
  -c, --columns int         number of columns rendered (default auto)
  -m, --magnification int   magnification factor (default 1)
  -f, --filename string     name of files to be saved (default "out")
```
Specific features:
- Julia fonts are built-in to secure broad international glyphs coverage.
- Reports back the number of columns actually used by the shot.

```text
╭info
╰Number of columns used: 62. Use '--columns' flag to impose it.
```

![example](example.png)

## Usage

Create *png* and *txt* color screenshots of the terminal command output.

```text
termos shot [shot flags] [--] command [command flags] [command arguments] [...] [flags]
```

After editing *txt* file, render *png* color screenshot of the file input.

```text
termos render [render flags] [--] filename
```

## Build

See [BUILD.md](BUILD.md) file.

## License

`termos` was created by Lukasz Lobocki. It is licensed under the terms of the CC0 v1.0 Universal license.

All components used retain their original licenses.

## Credits

`termos` is very heavily inspired by [homeport/termshot](https://github.com/homeport/termshot).

`termos` was created with [cookiecutter](https://cookiecutter.readthedocs.io/en/latest/) and [template](https://github.com/lukasz-lobocki/go-cookiecutter).
