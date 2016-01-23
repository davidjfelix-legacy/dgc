# dgc - docker garbage collector

DGC is a command line tool for cleaning up after docker.
It is modeled heavily after spotify's docker-gc project but was written in go using go-dockerclient.
The use of go and go-dockerclient allows this gc to be run remotely and make use of the docker remote api.
It can (not verified) be run from an Go-supporting OS and be run in a scratch docker container more easily than a bash-based solution.

TODO:

* Have A timed "cron" mode
* Daemonize process
* Allow multiple machines to be specified and run concurrently
