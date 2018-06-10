# rdocker

A tool to control a docker daemon on remote machines over an ssh tunnel.

## Example

```
RD_HOST=192.168.0.10:22 RD_USER=dev rdocker run hello-world
```

## Configuration

rdocker uses environment variables for configuration:

- *RD_HOST* - The remote host with which to establish a tunnel (REQUIRED)
- *RD_USER* - The user with which to establish a tunnel (default "root")
- *RD_PRIV_KEY* - The SSH private key used to establish a tunnel
- *RD_PRIV_KEY_FILE* - The path to the SSH private key used to establish a tunnel (default "~/.ssh/id_rsa")

If both *RD_PRIV_KEY* and *RD_PRIV_KEY_FILE* are set, *RD_PRIV_KEY* is used.