# httpmon

[![CircleCI](https://circleci.com/gh/joshuarubin/httpmon.svg?style=svg)](https://circleci.com/gh/joshuarubin/httpmon) [![GoDoc](https://godoc.org/jrubin.io/httpmon?status.svg)](https://godoc.org/jrubin.io/httpmon) [![Go Report Card](https://goreportcard.com/badge/jrubin.io/httpmon)](https://goreportcard.com/report/jrubin.io/httpmon) [![codecov](https://codecov.io/gh/joshuarubin/httpmon/branch/master/graph/badge.svg)](https://codecov.io/gh/joshuarubin/httpmon)

```sh
# httpmon -h
Usage of ./httpmon:
  -alert-duration duration
        the request rate must exceed alert-rate for this much time, on average, before an alert is triggered (default 2m0s)
  -alert-rate float
        number of requests/s for an alert to be triggered (default 10)
  -file string
        file to read, "-" for stdin (default "/var/log/access.log")
  -no-cat-log
        don't show the log lines
```
