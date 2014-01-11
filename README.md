rtot
====

- **R**un
- **T**his
- **O**ver
- **T**here

Another async remote execution thing because that's just what the world
needs, right?

## No really

Jobs live in memory and they aren't automatically garbage collected.
Jobs run as the same user running the `rtot-server`.  SSL isn't built
in, so if you're feeling paranoid you should probably put this behind
nginx or whatever.

## Example usage

Run the server somewhere:

``` bash
rtot-server -a=':8457' -s='supersecret'
```

Hit the server over HTTP:

``` bash
curl -H 'Rtot-Secret: supersecret' \
  -d 'echo wat is happening' \
  http://other-server.example.com:8457
```

which returns the path to the job id:

``` bash
/1
```

Check back on that job:

``` bash
curl -H 'Rtot-Secret: supersecret' \
  http://other-server.example.com:8457/1
```

which returns JSON including the job's stdout, stderr, and the exit
error if any:

``` javascript
{"out":"wat is happening\n","err":"","exit":null}
```

### a note on shebangs

If the data POSTed to the server does not start with `#!`, a shebang
of `#!/bin/bash` is prepended, otherwise it's assumed that the shebang
provided will be understood by the kernel.
