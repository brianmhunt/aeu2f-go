//
// AppEngine Universal 2 Factor
// (aeutf)
//
// License: MIT
//
package aeu2f

import (
  "fmt"
  "log"
  "time"
  "testing"
	"encoding/base64"

  "appengine/aetest"
  "appengine/datastore"

	"github.com/tstranex/u2f"
)


// Pad the input string to 4 bytes with "=".
func decodeBase64(s string) ([]byte, error) {
	for i := 0; i < len(s) % 4; i++ {
		s += "="
	}
	return base64.URLEncoding.DecodeString(s)
}


//
// Example data from https://developers.yubico.com/python-u2flib-server/
//
const fakeChallengeB64 = "RHlj0gKpjW-lbxeP6kSESNGlg2urIdbfYnqKAh7Hxlo"
var fakeChallenge, err = decodeBase64(fakeChallengeB64)
// var fakeHost = "http://localhost:8081"
var fakeHost = "https://aeu2f-demo.appspot.com"
var fakeRegistrationChallenge = u2f.Challenge{
  Timestamp: time.Now(),
  AppID: fakeHost,
  TrustedFacets: []string{fakeHost},
  Challenge: fakeChallenge }


var fakeRegistrationResponse2 = u2f.RegisterResponse{
  ClientData: "eyJvcmlnaW4iOiAiaHR0cDovL2xvY2FsaG9zdDo4MDgxIiwgImNoYWxsZW5nZSI6ICJEMnB6VFBaYTdicTY5QUJ1aUdRSUxvOXpjc1RVUlAyNlJMaWZUeUNraWxjIiwgInR5cCI6ICJuYXZpZ2F0b3IuaWQuZmluaXNoRW5yb2xsbWVudCJ9",
  RegistrationData: "BQSivQtJ6-lAgZ2qQ0aUGLEiJSRoLWUSGcmMO8C-GuibA0-xTvmuQfTqKyFJZWOUjGzEIgF4xV6gJ6itcagsyuUWQEQh9noDSu-WtzTOMhK_lKHxwHtQgJHCkzs4mukfpf310K5Dq9k6zBNtZ2RMBWgJhI7hJo4JiFn3k2GUNLwKZpwwggGHMIIBLqADAgECAgkAmb7osQyi7BwwCQYHKoZIzj0EATAhMR8wHQYDVQQDDBZZdWJpY28gVTJGIFNvZnQgRGV2aWNlMB4XDTEzMDcxNzE0MjEwM1oXDTE2MDcxNjE0MjEwM1owITEfMB0GA1UEAwwWWXViaWNvIFUyRiBTb2Z0IERldmljZTBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABDvhl91zfpg9n7DeCedcQ8gGXUnemiXoi-JEAxz-EIhkVsMPAyzhtJZ4V3CqMZ-MOUgICt2aMxacMX9cIa8dgS2jUDBOMB0GA1UdDgQWBBQNqL-TV04iaO6mS5tjGE6ShfexnjAfBgNVHSMEGDAWgBQNqL-TV04iaO6mS5tjGE6ShfexnjAMBgNVHRMEBTADAQH_MAkGByqGSM49BAEDSAAwRQIgXJWZdbvOWdhVaG7IJtn44o21Kmi8EHsDk4cAfnZ0r38CIQD6ZPi3Pl4lXxbY7BXFyrpkiOvCpdyNdLLYbSTbvIBQOTBFAiEA1uwJKNez6_BHdA2d-DPmRFJj19biYNkhN86SFH5Z_lYCICld2L3ZAVsm_uNFRt13_N9dlhGu50pb1ql8-_3_p5v1" }


// From client
var fakeRegistrationResponse = u2f.RegisterResponse{
  RegistrationData: "BQRM6zTJH7HRlC3yR4nO25ibCNXNRCsiyVjK6T1xa67lvbSDdzjvcvNoSW8pllLf6DWWX5j-7oTOYXdSiATuJ8K0QJNwNOkQqIBLOLtFxEBs6rtKiUc1D6rrGyexWCKsxElDFgPkvYyG88Vzjfej2dqYFEjHVTvLc4GRnZObENSe3tkwggJEMIIBLqADAgECAgR4wN8OMAsGCSqGSIb3DQEBCzAuMSwwKgYDVQQDEyNZdWJpY28gVTJGIFJvb3QgQ0EgU2VyaWFsIDQ1NzIwMDYzMTAgFw0xNDA4MDEwMDAwMDBaGA8yMDUwMDkwNDAwMDAwMFowKjEoMCYGA1UEAwwfWXViaWNvIFUyRiBFRSBTZXJpYWwgMjAyNTkwNTkzNDBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABLW4cVyD_f4OoVxFd6yFjfSMF2_eh53K9Lg9QNMg8m-t5iX89_XIr9g1GPjbniHsCDsYRYDHF-xKRwuWim-6P2-jOzA5MCIGCSsGAQQBgsQKAgQVMS4zLjYuMS40LjEuNDE0ODIuMS4xMBMGCysGAQQBguUcAgEBBAQDAgUgMAsGCSqGSIb3DQEBCwOCAQEAPvar9kqRawv5lJON3JU04FRAAmhWeKcsQ6er5l2QZf9h9FHOijru2GaJ0ZC5UK8AelTRMe7wb-JrTqe7PjK3kgWl36dgBDRT40r4RMN81KhfjFwthw4KKLK37UQCQf2zeSsgdrDhivqbQy7u_CZYugkFxBskqTxuyLum1W8z6NZT189r1QFUVaJll0D33MUcwDFgnNA-ps3pOZ7KCHYykHY_tMjQD1aQaaElSQBq67BqIaIU5JmYN7Qp6B1-VtM6VJLdOhYcgpOVQIGqfu90nDpWPb3X26OVzEc-RGltQZGFwkN6yDrAZMHL5HIn_3obd8fV6gw2fUX2ML2ZjVmybjBGAiEA_V8dGq-W1WwEO4LM8VEsNWAeK6GjxCC1ihqHW_K9H74CIQCcouyEm3V9DlqmOeJbe7qyki6f30qkiUfEBXVAAmolJg",
  ClientData: "eyJ0eXAiOiJuYXZpZ2F0b3IuaWQuZmluaXNoRW5yb2xsbWVudCIsImNoYWxsZW5nZSI6IlJIbGowZ0twalctbGJ4ZVA2a1NFU05HbGcydXJJZGJmWW5xS0FoN0h4bG8iLCJvcmlnaW4iOiJodHRwczovL2FldTJmLWRlbW8uYXBwc3BvdC5jb20iLCJjaWRfcHVia2V5Ijp7ImNydiI6IlAtMjU2Iiwia3R5IjoiRUMiLCJ4Ijoib0RxWGxjNEhYY2tvTDFxWnMxbTlIWEdvVllKTHB1d3FCUzJFWnJZaXBqOCIsInkiOiI1b2xZNlJYalBXOWhrUXoyX0dLckd4dGFHYjRmN1Y0aUZVYVdxQm1EaVFzIn19"}



//
// func TestDatabaseIdempotency(t *testing.T) {
//   // Ensure what we put in is what we get out.
//   key := datastore.NewKey(ctx, "Challenge", 0, 0, nil)
//   _, err = datastore.Put(ctx, key, &fakeRegistrationChallenge)
// 	if err != nil {
// 		t.Fatalf("datastore.Put error: %v", err)
// 	}
//   regCopy = new(u2f.Challenge)
//
//   // Note that the datastore rounds Timestamp to microsecond precision
//
// }


func TestNewChallenge(t *testing.T) {
  log.Printf("--- challenge ---")

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


func TestGoodRegistration(t *testing.T) {
  ctx, err := aetest.NewContext(nil)
  if err != nil {
    t.Fatal(err)
  }
  defer ctx.Close()

  var testId = "test-id-🔒"

  // Mimic NewChallenge
  ckey := makeKey(ctx, testId, "Challenge")
  _, err = datastore.Put(ctx, ckey, &fakeRegistrationChallenge)
	if err != nil {
		t.Fatalf("datastore.Put error: %v", err)
	}
  // log.Printf("Challenge: %+v", fakeRegistrationChallenge)

  if err := StoreResponse(ctx, testId, fakeRegistrationResponse); err != nil {
    t.Fatalf("StoreRegistration: %v", err)
  }

  // Load what was just saved and verify it.
  k := makeKey(ctx, testId, "Registration")
  var regi Registration
  if err := datastore.Get(ctx, k, &regi); err != nil {
    t.Fatalf("datastore.Get (%+v): %+v", k, regi)
  }

  // Verify the stored info.
  if regi.Counter != 0 {
    t.Error("Expected Counter to be 0")
  }

  if regi.UserIdentity != testId {
    t.Error("Expected user identity %v to be %v", regi.UserIdentity,
      testId)
  }

  u2fReg := new(u2f.Registration)
  if err := u2fReg.UnmarshalBinary(regi.U2FRegistrationBytes); err != nil {
    t.Fatalf("Unable to convert registration to struct: %+v", err)
  }
}
