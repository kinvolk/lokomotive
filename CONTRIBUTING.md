# Contributing

## Developer Certificate of Origin

By contributing, you agree to the Linux Foundation's Developer Certificate of Origin ([DCO](DCO)). The DCO is a statement that you, the contributor, have the legal right to make your contribution and understand the contribution will be distributed as part of this project.

## Commit guidelines

The title of the commit message should describe the _what_ about the
changes. Additionally, it is helpful to add the _why_ in the body of
the commit changes. Make sure to include any assumptions that you
might have made in this commit. Changes to unrelated parts of the
codebase should be kept as separate commits.

### Commit Format

```
<area>: <description of changes>

Detailed information about the commit message goes here
```

The title should not exceed 80 chars, although keeping it under 72
chars is appreciated.

Please wrap the body of commit message at a
maximum of 80 chars.

Here are a few example commit messages:

Good:
```
packet/flatcar-linux: Add support for reservation_ids on workers


```

Bad:
```
Add support for reservation_ids on workers for packet
```


Acceptable:
```
packet: Add support for reservation_ids on workers
```

This format is acceptable as sometimes nesting parts of the codebase
in the title can take up a lot of characters. Also at the same time,
using `packet/flatcar-linux` is redundant in the title, unless there
is another directory of the same name but with a different parent
directory, for eg: `docs/packet`.
