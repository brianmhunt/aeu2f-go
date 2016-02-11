//
// AppEngine Universal 2 Factor
// (aeutf)
//
// License: MIT
//
package aeu2f

import (
  "fmt"
  "testing"

  "appengine/aetest"
  "appengine/datastore"
)


func TestNewChallenge(t *testing.T) {
  ctx, err := aetest.NewContext(nil)

  if err != nil {
    t.Fatal(err)
  }

  // Create new challenge
  AppID = "tnc-appid"
  c, err := NewChallenge(ctx, "test")
  if err != nil {
    t.Fatal(err)
  }

  // Verify the challenge response
  if fmt.Sprintf("%T", c.Version) != "string" {
    t.Error("Expected c.Version to be a string.")
  }

  if fmt.Sprintf("%T", c.Challenge) != "string" {
    t.Error("Expected c.Challenge to be a string.")
  }

  if c.AppID != "tnc-appid" {
    t.Error("Expected c.AppID to be set appropriately.")
  }

  // Test that we've one item in the database, and that it stores a
  // u2f.Challenge
  q := datastore.NewQuery("Challenge")
  count, err := q.Count(ctx);
  if err != nil {
    t.Fatal(err)
  }
  if count != 1 {
    t.Fatalf("Expected one Challenge in the datastore, got %v.", count)
  }
}
