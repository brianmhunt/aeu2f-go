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
	"time"

	"appengine"
	"appengine/datastore"

	"github.com/tstranex/u2f"
)

// type Challenge struct {
// 	userIdentity string
// 	challenge u2f.Challenge
// }
//

// Registration stores the response to a registration challenge.
type Registration struct {
	UserIdentity string
	U2FRegistrationBytes []byte
	Counter uint32
	Created time.Time
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
func makeKey(ctx appengine.Context, stringKey, kind string) *datastore.Key {	parent := MakeParentKey(ctx)
	return datastore.NewKey(ctx, kind, stringKey, 0, parent)
}


// NewRegistrationChallenge creates a new U2F challenge and stores it in the
// datastore.
//
// Encode the response with e.g.
// 	 json.NewEncoder(w).Encode(req)
//
func NewRegistrationChallenge(ctx appengine.Context, userIdentity string) (*u2f.RegisterRequest, error) {
	// Generate a challenge
	c, err := u2f.NewChallenge(AppID, TrustedFacets); if err != nil {
		return nil, fmt.Errorf("u2f.NewChallenge error: %v", err)
	}

	// Save challenge to database.
	ckey := makeKey(ctx, userIdentity, "Challenge")
	if _, err := datastore.Put(ctx, ckey, c); err != nil {
		return nil, fmt.Errorf("datastore.Put error: %v", err)
	}

	// Return challenge request
	req := c.RegisterRequest()
	log.Printf("🍁  New Registration Challenge for %v: %+v [%+v]",
		userIdentity, req, ckey)
	return req, nil
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

	reg, err := u2f.Register(resp, challenge, &u2f.Config{true})
	if err != nil {
		return fmt.Errorf("u2f.Register error: %v", err)
	}

	buf, err := reg.MarshalBinary()
	if err != nil {
		return fmt.Errorf("reg.MarshalBinary error: %v", err)
	}

	// Save the registration in the datastore
	regi := Registration{UserIdentity: userIdentity, Counter: 0, U2FRegistrationBytes: buf, Created: time.Now()}
	// We set the stringKey to 0, because the user identity is not part of the
	// key.  We look up registrations by a datastore query, since there might
	// be multiple.
	k := makeKey(ctx, "", "Registration")
  if _, err := datastore.Put(ctx, k, &regi); err != nil {
    return fmt.Errorf("datastore.Put error: %v", err)
	}

	log.Printf("🍁  Registered: %+v [%+v]", userIdentity, k)

	return nil
}
