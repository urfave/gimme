# Releasing urfave/gimme

Releasing small batches often is [backed by
research](https://itrevolution.com/accelerate-book/) as part of the
virtuous cycles that keep teams and products healthy.

To that end, the overall goal of the release process is to send
changes out into the world as close to the time the commits were
merged to the `main` branch as possible. In this way, the community
of humans depending on this library are able to make use of the
changes they need **quickly**, which means they shouldn't have to
maintain long-lived forks of the project, which means they can get
back to focusing on the work on which they want to focus. This also
means that the @urfave/gimme team should be able to focus on
delivering a steadily improving product with significantly eased
ability to associate bugs and regressions with specific releases.

## Process

- Release versions follow [semantic versioning](https://semver.org/)
- Releases are associated with **signed, annotated git tags**[^1].
- Release notes are **automatically generated**[^1].

In the `main` or `v2-maint` branch, the current version is always
available via:

```sh
git describe --always --dirty --tags
```

**NOTE**: if the version reported contains `-dirty`, this is
indicative of a "dirty" work tree, which is not a great state for
creating a new release tag. Seek help from @urfave/gimme teammates.

For example, given a described version of `v1.7.1-7-g65c7203` and a
diff of `v1.7.1...` that contains only bug fixes, the next version
should be `v1.7.2`:

- Update `GIMME_VERSION` in [gimme](https://github.com/urfave/gimme/blob/main/gimme) via pull request.

- Once that is approved and merged make a tag locally:
```sh
TAG_VERSION=v1.7.2 make tag
git push origin v1.7.2
```

- Open the [the new release page](https://github.com/urfave/gimme/releases/new)
- At the top of the form, click on the `Choose a tag` select control and select `v1.7.2`
- In the `Write` tab below, click the `Auto-generate release notes` button
- At the bottom of the form, click the `Publish release` button
- :white_check_mark: you're done!

[^1]: This was not always true. There are many **lightweight git
  tags** present in the repository history. And releases with the wrong version
  number in source code.
