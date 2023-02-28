package oid

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/bsoncodec"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var testID = "5d6f6ff1646327ce31968d93"
var testIDSecs int64 = 1567584241 // Wednesday, September 4, 2019 8:04:01 AM
var testIDMachine uint32 = 6578983
var testIDPid uint16 = 52785
var testIDCounter int32 = 9866643

type test struct {
	ID primitive.ObjectID `bson:"_id"`
	V  ObjectID           `bson:"v,omitempty"`
	P  *ObjectID          `bson:"p,omitempty"`
	S  string             `bson:"s"`
	B  []byte             `bson:"b"`
	N  int32              `bson:"n,omitempty"`
}

func TestObjectID(t *testing.T) {
	tearUp(t, func(ctx context.Context, e *mongo.Collection) {
		id := NewObjectID()
		pID := NewObjectID()

		te := &test{ID: primitive.NewObjectID(), V: id, P: &pID}

		t.Run("marshal", func(t *testing.T) {
			if _, err := e.InsertOne(ctx, te); err != nil {
				t.Fatalf("expected nil, got %s", err)
			}

			res := e.FindOne(ctx, bson.M{"v": id})
			if err := res.Err(); err != nil {
				t.Fatalf("expected nil, got %s", err)
			}
		})
		t.Run("unmarshal", func(t *testing.T) {
			var out test
			res := e.FindOne(ctx, bson.M{"v": id})
			if err := res.Err(); err != nil {
				t.Fatalf("expected nil, got %v", err)
			}

			if err := res.Decode(&out); err != nil {
				t.Fatalf("expected nil, got %v", err)
			}

			expected := test{
				V:  id,
				ID: te.ID,
				P:  &pID,
			}

			if !reflect.DeepEqual(expected, out) {
				t.Fatalf("\nexpected: %+v \n got %+v", expected, out)
			}
		})
	})

	t.Run("bad_id", func(t *testing.T) {
		tearUp(t, func(ctx context.Context, e *mongo.Collection) {
			b := test{
				ID: primitive.NewObjectID(),
				V:  ObjectID("123"),
			}
			expected := "cannot transform type oid.test to a BSON Document: ObjectID(\"313233\") is not an ObjectID"
			if _, err := e.InsertOne(ctx, b); err.Error() != expected {
				t.Fatalf("expected %s\n, got %s", expected, err)
			}
		})
	})

	t.Run("id_string", func(t *testing.T) {
		tearUp(t, func(ctx context.Context, e *mongo.Collection) {
			b := test{
				ID: primitive.NewObjectID(),
				S:  testID,
			}

			if _, err := e.InsertOne(ctx, b); err != nil {
				t.Fatalf("expected nil, got %s", err)
			}

			type resp struct {
				S ObjectID `bson:"s"`
			}

			var out resp

			res := e.FindOne(ctx, bson.M{"s": testID})
			if err := res.Err(); err != nil {
				t.Fatalf("expected nil, got %s", err)
			}
			if err := res.Decode(&out); err != nil {
				t.Fatalf("expected nil, got %v", err)
			}
			expected, _ := primitive.ObjectIDFromHex(testID)
			if expected.String() != out.S.String() {
				t.Fatalf("\nexpected: %+v \n got %+v", expected, out.S)
			}
		})
	})

	t.Run("id_string_invalid", func(t *testing.T) {
		tearUp(t, func(ctx context.Context, e *mongo.Collection) {
			invalidString := "AsDvLlsldffpwoerpsdlfksld!!!fkds"
			b := test{
				ID: primitive.NewObjectID(),
				S:  invalidString,
			}

			if _, err := e.InsertOne(ctx, b); err != nil {
				t.Fatalf("expected nil, got %s", err)
			}

			type resp struct {
				S ObjectID `bson:"s"`
			}

			var out resp

			res := e.FindOne(ctx, bson.M{"s": invalidString})
			if err := res.Err(); err != nil {
				t.Fatalf("expected nil, got %s", err)
			}
			expected := fmt.Sprintf("error occurred while trying to convert, reason: invalid input to ObjectIDHex: \"%s\"", invalidString)
			err := res.Decode(&out).(*bsoncodec.DecodeError)
			if nil == err {
				t.Fatalf("expected error, got nil")
			}
			if err.Unwrap().Error() != expected {
				t.Fatalf("expected: %s. got %s", expected, err.Unwrap().Error())
			}
		})
	})

	t.Run("id_num", func(t *testing.T) {
		tearUp(t, func(ctx context.Context, e *mongo.Collection) {
			b := test{
				ID: primitive.NewObjectID(),
				N:  123,
			}

			if _, err := e.InsertOne(ctx, b); err != nil {
				t.Fatalf("expected nil, got %s", err)
			}

			type resp struct {
				N ObjectID `bson:"n"`
			}
			var out resp
			expected := "type 32-bit integer cannot be converted to objectID"
			res := e.FindOne(ctx, bson.M{"n": 123})
			if err := res.Err(); err != nil {
				t.Fatalf("expected nil, got %s", err)
			}
			err := res.Decode(&out).(*bsoncodec.DecodeError)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}

			if err.Unwrap().Error() != expected {
				t.Fatalf("expected %s, got %s", expected, err.Unwrap().Error())
			}
		})
	})
}

func TestObjectIDHex(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		id, err := ObjectIDHex(testID)
		if err != nil {
			t.Fatalf("expected nil got %v", nil)
		}

		if !id.Valid() {
			t.Fatalf("expected valid, got %v", id)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		id, err := ObjectIDHex("1234")
		expected := errors.New("invalid input to ObjectIDHex: \"1234\"")
		if err.Error() != expected.Error() {
			t.Fatalf("expected %v got %v", expected, err)
		}
		if id.Valid() {
			t.Fatalf("expected invalid, got %v", id)
		}
	})
}

func TestIsObjectIDHex(t *testing.T) {
	t.Run("false", func(t *testing.T) {
		if IsObjectIDHex("1234") {
			t.Fatalf("expected false, got true")
		}
	})

	t.Run("true", func(t *testing.T) {
		if !IsObjectIDHex(testID) {
			t.Fatalf("expected true, got false")
		}
	})
}

func TestStringRep(t *testing.T) {
	id, err := ObjectIDHex(testID)
	if err != nil {
		t.Fatalf("could not make objectId %v", err)
	}
	expected := "ObjectID(\"" + testID + "\")"

	if expected != id.String() {
		t.Fatalf("expected %s, got %s", expected, id.String())
	}
}

func TestHex(t *testing.T) {
	id, err := ObjectIDHex(testID)
	if err != nil {
		t.Fatalf("could not make objectId %v", err)
	}

	if testID != id.Hex() {
		t.Fatalf("expected %s, got %s", testID, id.Hex())
	}
}

func TestTime(t *testing.T) {
	id, err := ObjectIDHex(testID)
	if err != nil {
		t.Fatalf("could not make objectId %v", err)
	}

	tValid := time.Unix(testIDSecs, 0)
	if tValid != id.Time() {
		t.Fatalf("could not retrieve proper time stamp from objectId")
	}
}

func TestMachine(t *testing.T) {
	id, err := ObjectIDHex(testID)
	if err != nil {
		t.Fatalf("could not make objectId %v", err)
	}

	b := id.Machine()

	// Machine value is 3 bytes, convert to int to compare.
	mVal := uint32(b[2]) | uint32(b[1])<<8 | uint32(b[0])<<16
	if mVal != testIDMachine {
		t.Fatalf("could not retrieve proper machine ID from objectId")
	}
}

func TestPid(t *testing.T) {
	id, err := ObjectIDHex(testID)
	if err != nil {
		t.Fatalf("could not make objectId %v", err)
	}

	if id.Pid() != testIDPid {
		t.Fatalf("could not retrieve proper PID from objectId")
	}
}

func TestCounter(t *testing.T) {
	id, err := ObjectIDHex(testID)
	if err != nil {
		t.Fatalf("could not make objectId %v", err)
	}

	if id.Counter() != testIDCounter {
		t.Fatalf("could not retrieve proper counter from objectId")
	}
}

func TestJSON(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		p := map[string]interface{}{
			"v": testID,
		}

		b, err := json.Marshal(p)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		var out test

		if err := json.Unmarshal(b, &out); err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("invalid_12_byte", func(t *testing.T) {
		p := map[string]interface{}{
			"v": map[string]interface{}{
				"$oid": 123,
			},
		}

		b, err := json.Marshal(p)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		var out test

		expected := "invalid ObjectID in JSON: {\"$oid\":123} (encoding/hex: invalid byte: U+007B '{')"
		if err := json.Unmarshal(b, &out); err.Error() != expected {
			t.Fatalf("expected %s, got %v", expected, err)
		}
	})

	t.Run("extended_JSON_success", func(t *testing.T) {
		p := map[string]interface{}{
			"v": map[string]interface{}{
				"$oid": testID,
			},
		}

		b, err := json.Marshal(p)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		var out test

		if err := json.Unmarshal(b, &out); err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("extended_JSON_invalid_field", func(t *testing.T) {
		p := map[string]interface{}{
			"v": map[string]interface{}{
				"$oid": 234123,
			},
		}

		b, err := json.Marshal(p)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		var out test

		expected := "not an extended JSON ObjectID"
		if err := json.Unmarshal(b, &out); err.Error() != expected {
			t.Fatalf("expected %s, got %v", expected, err)
		}
	})

	t.Run("extended_JSON_invalid_key", func(t *testing.T) {
		p := map[string]interface{}{
			"v": map[string]interface{}{
				"fail": testID,
			},
		}

		b, err := json.Marshal(p)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		var out test

		expected := "not an extended JSON ObjectID"
		if err := json.Unmarshal(b, &out); err.Error() != expected {
			t.Fatalf("expected %s, got %v", expected, err)
		}
	})

	t.Run("extended_JSON_invalid_top_level", func(t *testing.T) {
		p := map[string]interface{}{
			"v": 123,
		}

		b, err := json.Marshal(p)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		var out test

		expected := "not an extended JSON ObjectID"
		if err := json.Unmarshal(b, &out); err.Error() != expected {
			t.Fatalf("expected %s, got %v", expected, err)
		}
	})

	t.Run("bad_length", func(t *testing.T) {
		p := map[string]interface{}{
			"v": "55",
		}

		b, err := json.Marshal(p)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		var out test

		expected := "invalid ObjectID in JSON: 55"
		if err := json.Unmarshal(b, &out); err.Error() != expected {
			t.Fatalf("expected %s, got %v", expected, err)
		}
	})

	t.Run("empty", func(t *testing.T) {
		p := map[string]interface{}{
			"v": "",
		}

		b, err := json.Marshal(p)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		var out test

		if err := json.Unmarshal(b, &out); err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if out.V.Valid() {
			t.Fatalf("expected empty id to be invalid, got %v", out)
		}
	})

	t.Run("hex", func(t *testing.T) {
		p := map[string]interface{}{
			"v": "xxxxxxxxxxxxxxxxxxxxxxxx",
		}

		b, err := json.Marshal(p)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		var out test

		expected := "invalid ObjectID in JSON: \"xxxxxxxxxxxxxxxxxxxxxxxx\" (encoding/hex: invalid byte: U+0078 'x')"
		if err := json.Unmarshal(b, &out); err.Error() != expected {
			t.Fatalf("expected %s, got %v", expected, err)
		}
	})

	t.Run("marshal", func(t *testing.T) {
		p := test{
			V: NewObjectID(),
		}

		b, err := json.Marshal(p)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		var out test
		if err := json.Unmarshal(b, &out); err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if out.V != p.V {
			t.Fatalf("expected equality, got %v and %v", out.V, p.V)
		}

		var outJ map[string]interface{}
		if err := json.Unmarshal(b, &outJ); err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if outJ["V"] != p.V.Hex() {
			t.Fatalf("expected %v, got %v", p.V.Hex(), outJ["V"])
		}

	})
}

func tearUp(t *testing.T, fn func(ctx context.Context, coll *mongo.Collection)) {
	mgoAddr := os.Getenv("MONGO_ADDR")
	if mgoAddr == "" {
		mgoAddr = "mongodb://localhost:27017"
	}
	ctx := context.Background()

	c, err := mongo.Connect(ctx, options.Client().ApplyURI(mgoAddr))
	if err != nil {
		t.Fatalf("Could not connect to mongo: %s", err)
	}

	defer c.Disconnect(ctx)

	rand.Seed(time.Now().UnixNano())
	DBName := fmt.Sprintf("testing-%d", rand.Intn(1000))
	db := c.Database(DBName)
	defer db.Drop(ctx)

	fn(ctx, db.Collection("test"))
}
