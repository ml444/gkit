# protoc-gen-go-validate

[![GoDoc](https://godoc.org/github.com/ml444/gkit/cmd/protoc-gen-go-validate?status.svg)](https://godoc.org/github.com/ml444/gkit/cmd/protoc-gen-go-validate)
[![Build Status](https://travis-ci.org/ml444/gkit.svg?branch=master)](https://travis-ci.org/ml444/gkit)
[![Go Report Card](https://goreportcard.com/badge/github.com/ml444/gkit)](https://goreportcard.com/report/github.com/ml444/gkit)
[![codecov](https://codecov.io/gh/ml444/gkit/branch/master/graph/badge.svg)](https://codecov.io/gh/ml444/gkit)


`protoc-gen-go-validate` 是一个用于生成 go 消息验证器的协议插件。
虽然协议缓冲区保证了结构化数据的类型，但它们无法强制执行值的语义规则检测。
该插件增加了对协议生成代码时对message字段值的约束验证。

**NOTE**: This is a fork of [bufbuild/protoc-gen-validate](https://github.com/bufbuild/protoc-gen-validate) with some modifications to only support Go.

> **变更1**: 只生成Go的校验代码。
> 
> **变更2**: 不包含校验规则的message不生成Validate代码。
> 
> **修改3**: 通用的变量(uuidPattern|emailPattern...)、和类型(ValidationError｜MultiError)抽离成公共部分，不再每个message自己维护，精简代码。
> 
> **修改4**: 只引用`google.golang.org/protobuf`，减少对其他包的依赖。
>
> **修改5**: 增加返回errorx.Error的错误（自定义服务的错误码）。
