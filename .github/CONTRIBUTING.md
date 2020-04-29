# Contributing

[tos-release]: https://help.github.com/articles/github-terms-of-service/#6-contributions-under-repository-license
[code-of-conduct]: CODE_OF_CONDUCT.md
[fork]: https://github.com/github/go-fault/fork
[pr]: https://github.com/github/go-fault/compare
[releases]: https://github.com/github/go-fault/releases

Hi there! We're thrilled that you'd like to contribute to this project. Your help is essential for keeping it great.

Contributions to this project are [released][tos-release] to the public under the [project's open source license](../LICENSE.md).

Please note that this project is released with a [Contributor Code of Conduct][code-of-conduct]. By participating in this project you agree to abide by its terms.

## Submitting a pull request

1. [Fork][fork] and clone the repository
1. Make sure the tests pass on your machine: `go test -race ./...`
1. Create a new branch: `git checkout -b my-branch-name`
1. Make your change, add tests, and make sure the tests still pass
1. Push to your fork and [submit a pull request][pr]
1. Pat your self on the back and wait for your pull request to be reviewed and merged.

Here are a few things you can do that will increase the likelihood of your pull request being accepted:

- Write tests that fully cover the code you've added.
- Write code that passes the linter: `golangci-lint run`
- Write code that maintains or decreases benchmarks: `go test -run=XXX -bench=. -benchmem`
- Keep your change as focused as possible. If there are multiple changes you would like to make that are not dependent upon each other, consider submitting them as separate pull requests.

## Releasing

This project follows standard semantic versioning. To release a new version create a release on the [releases page][releases].

## Resources

- [How to Contribute to Open Source](https://opensource.guide/how-to-contribute/)
- [Using Pull Requests](https://help.github.com/articles/about-pull-requests/)
- [GitHub Help](https://help.github.com)
