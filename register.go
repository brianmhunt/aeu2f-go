//
// Package aeu2f provides U2F stored in a database.
//
// AppEngine Universal 2 Factor
// (aeutf)
//
// License: MIT
//
package aeu2f

import (
	"log"
	"fmt"

	"appengine"
	"appengine/datastore"

	"github.com/tstranex/u2f"

	"encoding/base64"
	"strings"
	"errors"
	"crypto/elliptic"
	"encoding/asn1"
	"crypto/subtle"
	"crypto/x509"
	"encoding/json"

)

// type Challenge struct {
// 	userIdentity string
// 	challenge u2f.Challenge
// }
//

// Registration stores the response to a registration challenge.
type Registration struct {
	userIdentity string
	registration []byte
	counter int
}

// AppID identifies this application.  Must be set to the hostname.
var AppID string

// TrustedFacets is the list of U2F trusted facets
var TrustedFacets []string

// ChallengeTimeout is the time within which a user must respond to a
// U2F registration challenge.
var ChallengeTimeout = 1000 * 60  // milliseconds


// MakeParentKey returns a Key to be used as the parent for the model, to
// enforce strong consistency.
func MakeParentKey(ctx appengine.Context) *datastore.Key {
	return datastore.NewKey(ctx, "U2F", "Registration", 0, nil)
}


// makeKey creates a key for a strongly consistent model.
func makeKey(ctx appengine.Context, userIdentity, kind string) *datastore.Key {	parent := MakeParentKey(ctx)
	return datastore.NewKey(ctx, kind, userIdentity, 0, parent)
}


// NewChallenge creates a new U2F challenge and stores it in the datastore.
//
// Encode the response with e.g.
// 	 json.NewEncoder(w).Encode(req)
//
func NewChallenge(ctx appengine.Context, userIdentity string) (*u2f.RegisterRequest, error) {
	// Generate a challenge
	c, err := u2f.NewChallenge(AppID, TrustedFacets); if err != nil {
		// log.Printf("u2f.NewChallenge error: %v", err)
		// http.Error(w, "error", http.StatusInternalServerError)
		return nil, fmt.Errorf("u2f.NewChallenge error: %v", err)
	}

	// Save challenge to database.
	ckey := makeKey(ctx, userIdentity, "Challenge")
	if _, err := datastore.Put(ctx, ckey, c); err != nil {
		return nil, fmt.Errorf("datastore.Put error: %v", err)
	}

	// Return challenge request
	req := c.RegisterRequest()
	// log.Printf("🍁  New Challenge: %+v [%+v]", req, k)
	return req, nil
}

// ----------
func decodeBase64(s string) ([]byte, error) {
	for i := 0; i < len(s)%4; i++ {
		s += "="
	}
	return base64.URLEncoding.DecodeString(s)
}

func encodeBase64(buf []byte) string {
	s := base64.URLEncoding.EncodeToString(buf)
	return strings.TrimRight(s, "=")
}


func parseRegistration(buf []byte) (*u2f.Registration, []byte, error) {
	if len(buf) < 1+65+1+1+1 {
		return nil, nil, errors.New("u2f: data is too short")
	}

	var r u2f.Registration
	r.Raw = buf

	if buf[0] != 0x05 {
		return nil, nil, errors.New("u2f: invalid reserved byte")
	}
	buf = buf[1:]

	x, y := elliptic.Unmarshal(elliptic.P256(), buf[:65])
	if x == nil {
		return nil, nil, errors.New("u2f: invalid public key")
	}
	r.PubKey.Curve = elliptic.P256()
	r.PubKey.X = x
	r.PubKey.Y = y
	buf = buf[65:]

	khLen := int(buf[0])
	buf = buf[1:]
	if len(buf) < khLen {
		return nil, nil, errors.New("u2f: invalid key handle")
	}
	r.KeyHandle = buf[:khLen]
	buf = buf[khLen:]

	// The length of the x509 cert isn't specified so it has to be inferred
	// by parsing. We can't use x509.ParseCertificate yet because it returns
	// an error if there are any trailing bytes. So parse raw asn1 as a
	// workaround to get the length.
	sig, err := asn1.Unmarshal(buf, &asn1.RawValue{})
	if err != nil {
		return nil, nil, err
	}

	buf = buf[:len(buf)-len(sig)]
	cert, err := x509.ParseCertificate(buf)
	if err != nil {
		return nil, nil, err
	}
	r.AttestationCert = cert

	return &r, sig, nil
}

func verifyClientData(clientData []byte, challenge u2f.Challenge) error {
	var cd u2f.ClientData
	if err := json.Unmarshal(clientData, &cd); err != nil {
		return err
	}

	foundFacetID := false
	for _, facetID := range challenge.TrustedFacets {
		if facetID == cd.Origin {
			foundFacetID = true
			break
		}
	}
	if !foundFacetID {
		return errors.New("u2f: untrusted facet id")
	}

	c := encodeBase64(challenge.Challenge)
	log.Printf("cd: %+v", cd)
	log.Printf("😡 Comparing\n...%+v\n...%+v", cd.Challenge, c)
	log.Printf("😡 Comparing\n...%+v\n...%+v", []byte(cd.Challenge), []byte(c))
	if len(c) != len(cd.Challenge) ||
		subtle.ConstantTimeCompare([]byte(c), []byte(cd.Challenge)) != 1 {
		return errors.New("u2f:👺 challenge does not match")
	}
	log.Printf("🍀  Registration passed")

	return nil
}

func fakeRegister(resp u2f.RegisterResponse, c u2f.Challenge) (*u2f.Registration, error) {
	regData, err := decodeBase64(resp.RegistrationData)
	if err != nil {
		return nil, err
	}

	clientData, err := decodeBase64(resp.ClientData)
	if err != nil {
		return nil, err
	}

	reg, _, err := parseRegistration(regData)
	if err != nil {
		return nil, err
	}

	// log.Printf("😇  Comparing\n CLIENT DATA: %+v to\nCHALLENGE %+v", clientData, c)
	if err := verifyClientData(clientData, c); err != nil {
		return nil, err
	}

	return reg, nil
}



// StoreResponse checks whether, based on the given information, the given
// U2F response has addressed the challenge.
//
// Get the RegisterResponse with e.g.
// 	if err := json.NewDecoder(r.Body).Decode(&regResp); err != nil {
// 		http.Error(w, "invalid response: "+err.Error(), http.StatusBadRequest)
// 		return
// 	}
func StoreResponse(ctx appengine.Context, userIdentity string, resp u2f.RegisterResponse) error {
	// Load the most recent challenge.
	ckey := makeKey(ctx, userIdentity, "Challenge")

	var challenge u2f.Challenge
  if err := datastore.Get(ctx, ckey, &challenge); err != nil {
    return fmt.Errorf("datastore.Get error: %v", err)
  }
	log.Printf("⭐️   challenge: %+v", challenge)

	// Register the challenge & response
	log.Printf("🚩 resp: %+v\n💝  challenge: %+v", resp, challenge)
	if _, err := fakeRegister(resp, challenge); err != nil {
		return fmt.Errorf("fakeRegister error: %v", err)
	}
	log.Print("☘  Matches! ")
	// return fmt.Errorf("Orchestrated STOP...")
	reg, err := u2f.Register(resp, challenge, &u2f.Config{true})
	if err != nil {
		return fmt.Errorf("u2f.Register error: %v", err)
	}

	buf, err := reg.MarshalBinary()
	if err != nil {
		return fmt.Errorf("reg.MarshalBinary error: %v", err)
	}

	// Save the registration in the datastore
	regi := new(Registration)
	regi.userIdentity = userIdentity
	regi.counter = 0
	regi.registration = buf
	k := makeKey(ctx, userIdentity, "Registration")
  if _, err := datastore.Put(ctx, k, reg); err != nil {
    return fmt.Errorf("datastore.Put error: %v", err)
	}

	return nil
}
