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


// makeKey creates a key for a strongly consistent model.
func makeKey(ctx appengine.Context, userIdentity, kind string) *datastore.Key {	parent := datastore.NewKey(ctx, "U2F", "Registration", 0, nil)
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
	log.Printf("New Challenge: %+v", req)
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
  challenge := new(u2f.Challenge)
  if err := datastore.Get(ctx, ckey, challenge); err != nil {
    return fmt.Errorf("datastore.Get error: %v", err)
  }

	if challenge == nil {
		return fmt.Errorf("No challenge found for user %v", userIdentity)
	}

	// Register the challenge & response
	reg, err := u2f.Register(resp, *challenge, nil)
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
