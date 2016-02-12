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

	"github.com/tstranex/u2f"
)


// Example 8.1 from the FIDO U2F v1.0 NFC 2015-05-14 document, at:
// https://fidoalliance.org/specs/fido-u2f-v1.0-nfc-bt-amendment-20150514
//         /fido-u2f-raw-message-formats.html#registration-example
const example81RegResp = "0504b174bc49c7ca254b70d2e5c207cee9cf174820ebd77ea3c65508c26da51b657c1cc6b952f8621697936482da0a6d3d3826a59095daf6cd7c03e2e60385d2f6d9402a552dfdb7477ed65fd84133f86196010b2215b57da75d315b7b9e8fe2e3925a6019551bab61d16591659cbaf00b4950f7abfe6660e2e006f76868b772d70c253082013c3081e4a003020102020a47901280001155957352300a06082a8648ce3d0403023017311530130603550403130c476e756262792050696c6f74301e170d3132303831343138323933325a170d3133303831343138323933325a3031312f302d0603550403132650696c6f74476e756262792d302e342e312d34373930313238303030313135353935373335323059301306072a8648ce3d020106082a8648ce3d030107034200048d617e65c9508e64bcc5673ac82a6799da3c1446682c258c463fffdf58dfd2fa3e6c378b53d795c4a4dffb4199edd7862f23abaf0203b4b8911ba0569994e101300a06082a8648ce3d0403020347003044022060cdb6061e9c22262d1aac1d96d8c70829b2366531dda268832cb836bcd30dfa0220631b1459f09e6330055722c8d89b7f48883b9089b88d60d1d9795902b30410df304502201471899bcc3987e62e8202c9b39c33c19033f7340352dba80fcab017db9230e402210082677d673d891933ade6f617e5dbde2e247e70423fd5ad7804a6d3d3961ef871"



func TestNewChallenge(t *testing.T) {
  ctx, err := aetest.NewContext(nil)
  if err != nil {
    t.Fatal(err)
  }
  defer ctx.Close()

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
  q := datastore.NewQuery("Challenge").Ancestor(MakeParentKey(ctx))
  count, err := q.Count(ctx);
  if err != nil {
    t.Fatal(err)
  }

  if count != 1 {
    t.Fatalf("Expected one Challenge in the datastore, got %v.", count)
  }
}


func TestBadRegistration(t *testing.T) {
  resp := new(u2f.Challenge)
  resp.Challenge = []byte("DsM4LYoNL5Q17yVzUywxvIwFxhvdlxPTiGr4iWMYEmE")
  resp.AppID = "tnc-appid"
  resp.TrustedFacets = []string{resp.AppID}
}
