# mkchangelog

mkchangelog is a tool to generate a Markdown change log from Git Logs.

## Installation

### Usage

#### Example

All examples assume that `mkchangelog` is run in the directory of a cloned Git
repository.

- To create a changelog for release named `3.0.0` that contains all commit messages
  from the Git tag v2.1.4 to the current HEAD:

  ```shell
  mkchangelog -n 3.0.0 v2.1.4
  ```

### From a Release

- Download a release from <https://github.com/simplesurance/mkchangelog/releases>
- Extract the .tar.xz archive via `tar xJf <filename>` and copy it into your `$PATH`

### Via go get

```shell
go get github.com:simplesurance/mkchangelog
```
