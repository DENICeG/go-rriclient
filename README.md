# RRI 3.0 Client

Send requests to RRI v3.0 and read response via interactive CLI.

## Installation

Install and build from repository:

```
go get -u github.com/DENICeG/go-rriclient
go install github.com/DENICeG/go-rriclient
```

## Usage

### Specify Connection and Credentials

The RRI address is defined by `host:port`. There are three ways to setup a connection to a RRI server:

1. **Server address only**. Will open an unauthenticated RRI session:

```
go-rriclient localhost:51131
```

2. **Server address and credentials**. Only pass user to interactively ask for password:

```
go-rriclient localhost:51131 -u DENIC-1000001-RRI -p secret
```

3. **Environment**. Will interactively prompt for address and credentials on first call:

```
go-rriclient -e {name}
```

4. **Interactively select environment**:

```
go-rriclient
```

The environment files are stored in `~/.rri-client`.

### RRI Command execution

You can interact with the rri-client in two modes. All modes can be combined with any of the previously described connection types:

1. **File source mode**

```
go-rriclient localhost:51131 -f examples/login_logout.txt
go-rriclient localhost:51131 -u DENIC-1000001-RRI -p secret -f examples/login_logout.txt
go-rriclient -e {name} -f examples/login_logout.txt
go-rriclient -f examples/login_logout.txt
```

This mode will execute all RRI commands contained in the specified file. Commands are separated by a line starting with `=-=`. When using an unauthenticated session, the file must also contain a login command to open an authenticated session.

2. **Interactive mode**

```
go-rriclient localhost:51131
go-rriclient localhost:51131 -u DENIC-1000001-RRI -p secret
go-rriclient -e {name}
go-rriclient
```

Will open a bash-like interactive CLI with command completion for convenient RRI access. See section *Interactive Commands* for a detailed explanation.

### Command Line Args

See the following list for a complete overview on all available flags:

| Flag | Short | Description |
| ---- | ----- | ----------- |
| `--user {username}` | `-u` | Optional RRI username to automatically log in. |
| `--pass {password}` | `-p` | Optional RRI user password to automatically log in. |
| `--file {file}` | `-f` | File containing RRI queries to process. |
| `--env {environment name}` | `-e` | Name of the environment to create or use. |
| `--verbose` | `-v` | Verbose mode for more detailed output. |

## Interactive Commands

When running in interactive mode, you can use the following commands:

| Command Name and Args | Description |
| --------------------- | ----------- |
| `login {user} {pass}` | Log in to a RRI account. |
| `logout` | Log out from the current RRI account. |
| `create domain {domain} {...}` | Send a CREATE command for a new domain. |
| `check domain {domain}` | Send a CHECK command for a specific domain. |
| `info {domain}` | Send an INFO command for a specific domain. |
| `update domain {domain} {...}` | Send an UPDATE command for a new domain. |
| `delete domain {domain}` | Send a DELETE command for a specific domain. |
| `restore {domain}` | Send a RESTORE command for a specific domain. |
| `create authinfo1 {domain} {secret}` | Send a CREATE-AUTHINFO1 command for a specific domain with auth info secret. |
| `create authinfo2 {domain}` | Send a CREATE-AUTHINFO2 command for a specific domain. |
| `chprov {domain} {secret} {...}` | Send a CHPROV command for a specific domain with auth info secret. |
| `file {path}` | Process a query file as accepted by flag `--file`. |
| `xml` | Toggle xml mode. |
| `verbose` | Toggle verbose mode. |
| `dry` | Toggle dry mode to only print out raw queries. |

### create

The full `create` command accepts the following parameters:

```
create domain {domain} {holder} {abuse-contact} {general-request} {nserver-1} {nserver-2} ...
```

The parameters `holder`, `abuse-contact`, `general-request` are handles. You can pass an arbitrary number of name servers at the end. An interactive prompt will be opened for all missing parameters.

### update

Behaves the same as `create`.

### chprov

The full `chprov` command is similar to the `create` command. It accepts the following parameters:

```
chprov {domain} {secret} {holder} {abuse-contact} {general-request} {nserver-1} {nserver-2} ...
```

It behaves exactly like the `create` command for missing parameters. You need to supply the same secret as passed to `authinfo1` to successfully take control over the domain.

## Thanks

Thanks to [sebidude](https://github.com/sebidude) for the [protocol implementation](https://github.com/sebidude/go-rri).