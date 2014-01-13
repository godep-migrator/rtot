rtot
====

**R**un **T**his **O**ver **T**here (<q>Arr-tot</q>)

Another async remote execution thing because that's just what the world
needs, right?

[![Build Status](https://travis-ci.org/modcloth-labs/rtot.png?branch=master)](https://travis-ci.org/modcloth-labs/rtot)

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
  http://other-server.example.com:8457/jobs
```

which returns the newly-created job, including its id and relative URL:

``` bash
{
  "jobs": [
    {
      "href": "/jobs/0",
      "id": 0,
      "state": "new",
      "start": "0001-01-01 00:00:00 +0000 UTC",
      "complete": "0001-01-01 00:00:00 +0000 UTC",
      "create": "2014-01-12 03:42:32.314152969 +0000 UTC"
    }
  ]
}
```

Check back on that job:

``` bash
curl -H 'Rtot-Secret: supersecret' \
  http://other-server.example.com:8457/jobs/0
```

which returns JSON including the job's stdout, stderr, and the exit
error if any:

``` javascript
{
  "jobs": [
    {
      "href": "/jobs/0",
      "id": 0,
      "out": "wat is happening\n",
      "state": "complete",
      "start": "2014-01-12 03:42:32.315039718 +0000 UTC",
      "complete": "2014-01-12 03:42:32.328346325 +0000 UTC",
      "create": "2014-01-12 03:42:32.314152969 +0000 UTC"
    }
  ]
}
```

Jobs that haven't exited yet will return with a state of `"running"` and
any stdout or stderr that have been collected so far:

``` javascript
{
  "jobs": [
    {
      "href": "/jobs/0",
      "id": 0,
      "out": "ready\nset\nwait for it\n",
      "state": "running",
      "start": "2014-01-12 03:46:38.297295634 +0000 UTC",
      "complete": "0001-01-01 00:00:00 +0000 UTC",
      "create": "2014-01-12 03:46:38.296604443 +0000 UTC"
    }
  ]
}
```

## A note on shebangs

If the data POSTed to the server does not start with `#!`, a shebang
of `#!/bin/bash` is prepended, otherwise it's assumed that the shebang
provided will be understood by the kernel.  The server *does not* try to
do anything fancy based on the content type of the request, and is sure
to offend purists.

## Job cleanup

As mentioned above, jobs are not automatically garbage collected.
Instead, it's up to you to clean up after yourself:

``` bash
curl -H 'Rtot-Secret: supersecret' \
  -X DELETE \
  http://other-server.example.com:8457/jobs/1
```

The response for a successful job delete will have a status of 204 and
no body.

## Death

Since rtot is all about arbitrary superpowers, it's also possible to
make it exit via the web, which is one way to restart it and/or purge
all jobs:

``` bash
curl -H 'Rtot-Secret: supersecret' \
  -X DELETE \
  http://other-server.example.com:8457/
```
