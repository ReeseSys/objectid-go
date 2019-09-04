# ObjectID-Go

ObjectID-Go helps to bridge the gap between the new [mongo-go-driver](https://github.com/mongodb/mongo-go-driver) and previous community run drivers such as [mgo](https://github.com/globalsign/mgo).

For people not migrating, this also helps cover some of the pitfalls and frustrations of mongo-go-driver's primitive package which is extremely error prone, leaks too much of the driver to the clients of an API, and can even cause unexpected panics by sending any invalid payload of 12 bytes.

This package follows the community run driver standard of using strings to represent objectIDs instead of [12]bytes, allowing for a much smoother development experience.

##Features

1. This package automatically converts all oid.ObjectID types into primitive.ObjectIDs when marshalling or unmarshalling into bson
2. no panics on JSON unmarshalling
3. uses string types to avoid the driver bleeding to the API
4. fixes vet errors that was rampant in the community drivers
5. combines the best features of community drivers and the primitive package
