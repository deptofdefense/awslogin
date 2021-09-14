# AWS Login Tool

This is a tool used to log into AWS accounts using 1Password MFA tokens.

## Prerequisites

Before using this tool you must install these prerequisites:

- [1Password Command Line Tool](https://support.1password.com/command-line-getting-started/)
- [GNU sed](https://formulae.brew.sh/formula/coreutils)

```sh
brew install coreutils
```

In your environment add:

```sh
PATH="$(brew --prefix)/opt/coreutils/libexec/gnubin:$PATH"
```

## Environment variables

```sh
op signin defense-digital-service.1password.com first.last@dds.mil --shorthand dds
```
