// Package oid provides an easy to use/integrate abstraction layer between your code and the primitive package of the official mongo driver
//
// helps to bridge the gap between the new mongo-go-driver and previous community run drivers such as the mgo package.
// For people not migrating, this also helps cover some of the pitfalls and frustrations of mongo-go-driver's primitive package which is extremely brittle, leaks too much of the driver to the clients of the API, and can even cause unexpected panics when unmarshalling
//
// This package follows the community run driver standard of using strings to represent objectIDs instead of [12]bytes, allowing for a much smoother development experience.
//
// Features
//
// 1. This package automatically unmarshalls all objectId strings in a JSON payload into oid.ObjectID types including support for mongos EXTJSON. And Un/Marshalls the oid.ObjectID types into primitive.ObjectIDs when interacting with bson
//
// 2. no panics on JSON unmarshalling
//
// 3. uses string types to avoid the driver bleeding to the API and give more control to the dev
//
// 4. fixes vet errors that was rampant in the community drivers
//
// 5. combines the best features of community drivers and the primitive package
//
// 6. makes migrating significantly easier
package oid

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

// ObjectID is a unique ID identifying a BSON value. It must be exactly 12 bytes
// long.
//
// it does not support mongo's extJSON spec
// http://www.mongodb.org/display/DOCS/Object+Ids
type ObjectID string

// ObjectIDHex returns an ObjectID from the provided hex representation.
// Calling this function with an invalid hex representation will
// return an error. See the IsObjectIDHex function.
func ObjectIDHex(s string) (ObjectID, error) {
	d, err := hex.DecodeString(s)
	if err != nil || len(d) != 12 {
		return ObjectID(d), fmt.Errorf("invalid input to ObjectIDHex: %q", s)
	}
	return ObjectID(d), nil
}

// IsObjectIDHex returns whether s is a valid hex representation of
// an ObjectID. See the ObjectIDHex function.
func IsObjectIDHex(s string) bool {
	_, err := primitive.ObjectIDFromHex(s)
	return err == nil
}

// NewObjectID returns a new unique ObjectID.
func NewObjectID() ObjectID {
	id, _ := ObjectIDHex(primitive.NewObjectID().Hex())
	return id
}

// String returns a hex string representation of the id.
// Example: ObjectIDHex("4d88e15b60f486e428412dc9").
func (id ObjectID) String() string {
	return fmt.Sprintf(`ObjectID("%x")`, string(id))
}

// Hex returns a hex representation of the ObjectID.
func (id ObjectID) Hex() string {
	return hex.EncodeToString([]byte(id))
}

// Valid confirms that the objectID is valid
func (id ObjectID) Valid() bool {
	_, err := primitive.ObjectIDFromHex(id.Hex())
	return err == nil
}

// byteSlice returns byte slice of id from start to end.
// Calling this function with an invalid id will cause a runtime panic.
func (id ObjectID) byteSlice(start, end int) []byte {
	if len(id) != 12 {
		panic(fmt.Sprintf("invalid ObjectID: %q", string(id)))
	}
	return []byte(string(id)[start:end])
}

// Time returns the timestamp part of the id.
// It's a runtime error to call this method with an invalid id.
func (id ObjectID) Time() time.Time {
	// First 4 bytes of ObjectID is 32-bit big-endian seconds from epoch.
	secs := int64(binary.BigEndian.Uint32(id.byteSlice(0, 4)))
	return time.Unix(secs, 0)
}

// Note: The ObjectID spec was changed in 2018: Machine ID and
// ProcessID were replaced by a single 5-byte random value. According
// to the spec, drivers MUST NOT have an accessor method on an ObjectID
// to obtain the random value.
// https://github.com/mongodb/specifications/blob/master/source/objectid.rst

// Deprecated: Machine returns the 3-byte machine id part of the id.
// It's a runtime error to call this method with an invalid id.
func (id ObjectID) Machine() []byte {
	return id.byteSlice(4, 7)
}

// Deprecated: Pid returns the process id part of the id.
// It's a runtime error to call this method with an invalid id.
func (id ObjectID) Pid() uint16 {
	return binary.BigEndian.Uint16(id.byteSlice(7, 9))
}

// Counter returns the incrementing value part of the id.
// It's a runtime error to call this method with an invalid id.
func (id ObjectID) Counter() int32 {
	b := id.byteSlice(9, 12)
	// Counter is stored as big-endian 3-byte value
	return int32(uint32(b[0])<<16 | uint32(b[1])<<8 | uint32(b[2]))
}

// MarshalBSONValue satisfies the decoding interface for the mongo driver
func (id ObjectID) MarshalBSONValue() (bsontype.Type, []byte, error) {
	objID, err := primitive.ObjectIDFromHex(id.Hex())
	if err != nil {
		return bsontype.ObjectID, []byte{}, fmt.Errorf("%s is not an ObjectID", id.String())
	}

	val := bsonx.ObjectID(objID)
	return val.MarshalBSONValue()
}

// UnmarshalBSONValue satisfies the decoding interface for the mongo driver
func (id *ObjectID) UnmarshalBSONValue(t bsontype.Type, b []byte) error {
	if t != bsontype.ObjectID && t != bsontype.String {
		return fmt.Errorf("type %s cannot be converted to %s", t, bsontype.ObjectID)
	}

	val := bsonx.Undefined()
	if err := val.UnmarshalBSONValue(t, b); err != nil {
		return fmt.Errorf("invalid objectID from source: %v", err)
	}

	var oid ObjectID
	var err error
	if t == bsontype.ObjectID {
		oid, err = ObjectIDHex(val.ObjectID().Hex())
	} else {
		oid, err = ObjectIDHex(val.String())
	}

	if nil != err {
		return fmt.Errorf("error occurred while trying to convert, reason: %s", err)
	}

	*id = oid

	return nil
}

// MarshalJSON turns a bson.ObjectID into a json.Marshaller.
func (id ObjectID) MarshalJSON() ([]byte, error) {
	return []byte("\"" + id.Hex() + "\""), nil
}

var nullBytes = []byte("null")

// UnmarshalJSON populates the byte slice with the ObjectID. If the byte slice is 64 bytes long, it
// will be populated with the hex representation of the ObjectID. If the byte slice is twelve bytes
// long, it will be populated with the BSON representation of the ObjectID. Otherwise, it will
// return an error.
func (id *ObjectID) UnmarshalJSON(b []byte) error {
	var buf [12]byte
	switch len(b) {
	case 12:
		_, err := hex.Decode(buf[:], b)
		if err != nil {
			return fmt.Errorf("invalid ObjectID in JSON: %s (%s)", string(b), err)
		}
	default:

		// Extended JSON
		var res interface{}
		if err := json.Unmarshal(b, &res); err != nil {
			return err
		}
		str, ok := res.(string)
		if !ok {
			m, ok := res.(map[string]interface{})
			if !ok {
				return errors.New("not an extended JSON ObjectID")
			}
			oid, ok := m["$oid"]
			if !ok {
				return errors.New("not an extended JSON ObjectID")
			}
			str, ok = oid.(string)
			if !ok {
				return errors.New("not an extended JSON ObjectID")
			}

		}

		if len(b) == 2 && b[0] == '"' && b[1] == '"' || bytes.Equal(b, nullBytes) {
			*id = ""
			return nil
		}

		if len(str) != 24 {
			return fmt.Errorf("invalid ObjectID in JSON: %s", str)
		}

		_, err := hex.Decode(buf[:], []byte(str))
		if err != nil {
			return fmt.Errorf("invalid ObjectID in JSON: %s (%s)", string(b), err)
		}

		*id = ObjectID(string(buf[:]))
	}

	return nil
}
