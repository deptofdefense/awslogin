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

You must then sign into 1Password at least once using the `op` command. Also include a `--shorthand` for future use.

```sh
op signin defense-digital-service.1password.com first.last@dds.mil --shorthand dds
```

For this to work you must have at least one `login` category entry in your 1Password vault. It needs a `one-time password` section as well
as a custom section named `ACCOUNT_INFO`. Additionally, one of the items in the section needs to be `Account Alias`.

## Example Usage

### Sign In to 1Password

```sh
go run github.com/deptofdefense/awslogin op-signin
```

This creates a json file in `~/.op_session` with the details for your 1Password session. This session expires after 30 minutes.

### Log Into AWS

```sh
go run github.com/deptofdefense/awslogin login
```

Follow the prompts which will look like:

```text
0 AWS alias1
1 AWS alias2

Choose a secret's number: 1

You chose: AWS alias2
Account Alias: alias2
MFA Token: 764417
```

Then your browser should open and log you into the AWS Console.
