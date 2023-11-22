# gkit
[![Build Status](https://travis-ci.org/go-gadgets/gkit.svg?branch=master)](https://travis-ci.org/go-gadgets/gkit)
[![Coverage Status](https://coveralls.io/repos/github/go-gadgets/gkit/badge.svg?branch=master)](https://coveralls.io/github/go-gadgets/gkit?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-gadgets/gkit)](https://goreportcard.com/report/github.com/go-gadgets/gkit)
[![GoDoc](https://godoc.org/github.com/go-gadgets/gkit?status.svg)](https://godoc.org/github.com/go-gadgets/gkit)

gkit is a microservice framework that integrates `HTTP` and `GRPC` communication protocols. 
It aims to be easy to use. By encapsulating the components of daily WEB development and 
using the [gctl](https://github.com/ml444/gctl) tool, the basic code of the project 
can be quickly generated, so that developers can Focus more on business logic development.

## Project structure introduction

```
.
├── dbx
├── errorx
├── listoption     
├── log
├── metrics
├── middleware
├── pkg
│   ├── auth
│   ├── env
│   ├── header
│   └── routine
├── transport
├── go.mod
└── go.sum
```
- **errorx**: error handling module
- **dbx**: Gorm-based database module
- **listoption**: Paging query module, which defines the conditional filtering method of list data, as well as the parameters and result return of paging query.
- **log**: log module
- **metrics**: indicator module
- **middleware**: middleware module
- **transport**: transport module
- **pkg**: public module, including some basic tool classes
- **internal**: internal module
