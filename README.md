gofuzzyourself
===

This tool is part of the learning process of the Go programming language.

What does it do
---
- GET fuzzing
- Header fuzzing (via `-H` argument)
- Follows redirects automatically (via `-follow` argument)
- Ability to set user agent (via `-user-agent` or `-H` arguments)

Filters
---
- Keyword filtering (via `-contains` argument)
- Hides or shows requests by status code (via `-hide` and `-show` arguments)

Getting & using
---
```bash
$ go get -u github.com/vlad-s/gofuzzyourself
$ gofuzzyourself -h
```