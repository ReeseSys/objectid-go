# ObjectID-Go

[![GoDoc](https://godoc.org/github.com/qhenkart/objectid-go?status.svg)](https://godoc.org/github.com/qhenkart/objectid-go)
[![CircleCI](https://circleci.com/gh/qhenkart/objectid-go.svg?style=svg)](https://circleci.com/gh/qhenkart/objectid-go)

ObjectID-Go helps to bridge the gap between the new [mongo-go-driver](https://github.com/mongodb/mongo-go-driver) and previous community run drivers such as [mgo](https://github.com/globalsign/mgo).

For people not migrating, this also helps cover some of the pitfalls and frustrations of mongo-go-driver's primitive package which is extremely brittle, leaks too much of the driver to the clients of the API, and can even cause unexpected panics when unmarshalling

This package follows the community run driver standard of using strings to represent objectIDs instead of [12]bytes, allowing for a much smoother development experience.

## Features

1. This package automatically converts all oid.ObjectID types into primitive.ObjectIDs when marshalling or unmarshalling into bson
2. no panics on JSON unmarshalling
3. uses string types to avoid the driver bleeding to the API and give more control to the dev
4. fixes vet errors that was rampant in the community drivers
5. combines the best features of community drivers and the primitive package
6. makes migrating significantly easier

## Install

`go get -u github.com/qhenkart/objectid-go`

or

`dep ensure --add github.com/qhenkart/objectid-go`

## Usage

```
type User struct {
  ID oid.ObjectID `bson:"_id" json:"id"`
}
```

and use the struct with json un/marshallers and the mongo-go-driver methods without a second thought

Check the godocs for additional functionality and helpers
