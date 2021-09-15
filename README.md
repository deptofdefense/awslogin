# AWS Login Tool

This is a tool used to log into AWS accounts using 1Password MFA tokens.

## Prerequisites

Before using this tool you must install these prerequisites:

- [1Password Command Line Tool](https://support.1password.com/command-line-getting-started/)
- [aws-vault](https://github.com/99designs/aws-vault)
- A web browser such as Chrome, Safari, or Firefox

You must then sign into 1Password at least once using the `op` command. Also include a `--shorthand` for future use.

```sh
op signin defense-digital-service.1password.com first.last@dds.mil --shorthand dds
```

For this to work you must have at least one `login` category entry in your 1Password vault. It needs a `one-time password` section as well
as a custom section named `ACCOUNT_INFO`. Additionally, one of the items in the section needs to be `ACCOUNT_ALIAS`. These can be configured
to the user's desire. As an example:

![1Password Login Example](./images/1password_login.png)

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

It is also possible to have a faster experience by filtering. If you know the alias in advance use this syntax:

```sh
go run github.com/deptofdefense/awslogin login known-alias
```

Follow the prompts which will look like:

```text
Account Alias: known-alias
MFA Token: 764418
```

The difference here is being directly logged in with no prompts. Multiple filters can be used if needed.

The browser to use can also be changed if desired:

```sh
go run github.com/deptofdefense/awslogin login known-alias --browser chrome
```
