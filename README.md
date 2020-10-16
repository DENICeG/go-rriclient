# DENIC RRI Client

This application is a client for the RRI with an interactive CLI developed by [DENIC eG](https://denic.de), to send requests and info commands to and receive responses from the DENIC registry server.

## Installation

The DENIC RRI client is developed in GO and can be installed and built from the repository:

```
go get -u github.com/DENICeG/go-rriclient
go install github.com/DENICeG/go-rriclient
```

## Connections and Credentials

The RRI address is defined by `host:port`. There are three ways to setup a connection to an RRI server:

1. **Server address only**

This command opens an unauthenticated RRI session:

```
go-rriclient localhost:51131
```

You have to type in `login {username}` to get access to the RRI server. This will trigger a prompt for your password.

2. **Server address and credentials**

You can log in directly with username and password added in the command line:

```
go-rriclient localhost:51131 -u {username} -p {password}
```

3. **Set up Environment Files**

If you have access to one or more RRI environments with different sets of server addresses and credential information, you can store this access information in an easy-to-use picklist with alias names. You add entries to the picklist with:

```
go-rriclient -e {alias name}
```

The DENIC RRI client will ask you for host, port, username and password that you would like to store with the alias. The information is stored in an environment file in `~/.rri-client`.

When you type in:

```
go-rriclient
```

the DENIC RRI client will start with the picklist.

## DENIC RRI Client Modes

You can interact with the DENIC RRI client in two modes. All modes can be combined with any of the previously described connection types. See sections *CLI Arguments*, *RRI Commands* and *RRI Request Examples* for a detailed explanation of CLI arguments and RRI commands/parameters.

1. **File Source Mode**

This mode will execute all RRI commands and arguments contained in the file you have specified (`example.txt`). Commands are separated by a line starting with a dash. When using an unauthenticated session, the file must also contain a login command to open an authenticated session.

```
go-rriclient localhost:51131 -f example.txt
go-rriclient localhost:51131 -u (username) -p (password) -f example.txt
go-rriclient -e {alias name} -f example.txt
go-rriclient -f example.txt
```
2. **Interactive Mode**

This mode will open a bash-like interactive CLI with command completion for convenient RRI access.

```
go-rriclient localhost:51131
go-rriclient localhost:51131 -u (username) -p (password)
go-rriclient -e {alias name}
go-rriclient
```

## CLI Arguments

| Flag | Short | Description |
| ---- | ----- | ----------- |
| `--user {username}` | `-u` | RRI username to log in. |
| `--pass {password}` | `-p` | RRI password to log in. |
| `--file {file}` | `-f` | File containing RRI queries to process. |
| `--env {alias name}` | `-e` | Name of the environment to create or use. |
| `--delete-env {alias name}` | | Delete an existing environment. |
| `--list-env` | | Display a list of all environments. |
| `--fail` | | Exit with code 1 if RRI returns a failed result. |
| `--verbose` | `-v` | Verbose mode for more detailed output. |
| `--insecure` | | Skip SSL certificate check to enable self signed certificates. |
| `--version` | | Print out the application version and exit. |
| `--dump-cli-config` | | Print out the application cli configuration and exit. |

## RRI Commands

You can use the following commands in file mode and interactive mode:

| Command and Parameter | Description |
| --------------------- | ----------- |
| `login {username} {password}` | Log in to a RRI account. |
| `logout` | Log out from the current RRI account. |
| `check handle {handle}` | Send a CHECK command for a specific handle. |
| `info handle {handle}` | Send an INFO command for a specific handle. |
| `create domain {domain} {...}` | Send a CREATE command for a new domain. |
| `check domain {domain}` | Send a CHECK command for a specific domain. |
| `info domain {domain}` | Send an INFO command for a specific domain. |
| `update domain {domain} {...}` | Send an UPDATE command for a new domain. |
| `delete domain {domain}` | Send a DELETE command for a specific domain. |
| `restore {domain}` | Send a RESTORE command for a specific domain. |
| `transit {domain}` | Send a TRANSIT command without disconnect for a specific domain. |
| `create authinfo1 {domain} {secret}` | Send a CREATE-AUTHINFO1 command for a specific domain with AuthInfo. |
| `create authinfo2 {domain}` | Send a CREATE-AUTHINFO2 command for a specific domain. |
| `chprov {domain} {secret} {...}` | Send a CHPROV command for a specific domain with AuthInfo. |
| `raw` | Enter a raw query and send to RRI. |
| `raw {command}` | Send a command like `"version: 3.0\naction: queue-read"`. |
| `file {path}` | Process a query file as accepted by flag `--file`. |
| `xml` | Toggle XML mode. **NOT implemented yet** |
| `verbose` | Toggle verbose mode. |
| `dry` | Toggle dry mode to only print out raw queries. |

## RRI Request Examples

**Create Domain/Update Domain**

The full `create domain` and `update domain` commands accept the following parameters:

```
create domain {domain} {holder} {general-request} {abuse-contact} {nserver-1} {nserver-2} ...
```

The parameters `holder`, `general-request` and `abuse-contact` are handles. You specify an arbitrary number of name servers at the end. An interactive prompt will be opened for all missing parameters.

**Chprov**

The `chprov` command is like the `create domain` command. It behaves exactly like the `create domain` command and accepts the following parameters:

```
chprov {domain} {secret} {holder} {general-request} {abuse-contact} {nserver-1} {nserver-2} ...
```

## RRI Package

This repository also provides the Go package `github.com/DENICeG/go-rriclient/pkg/rri` that can be used as base for custom implementations. See [github.com/DENICeG/go-rriclient/tree/master/pkg/rri](https://github.com/DENICeG/go-rriclient/tree/master/pkg/rri) for a detailed, technical explanation and usage examples.

## Thanks

Thanks to [sebidude](https://github.com/sebidude) for the [protocol implementation](https://github.com/sebidude/go-rri).
