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
	"errors"

	"appengine"
	"appengine/datastore"

	"github.com/tstranex/u2f"
)

type Challenge struct {
	user_identity string
	challenge u2f.Challenge
}

type Registration struct {
	user_identity string
	registration []byte
	counter int
}


var AppID string
var TrustedFacets []string
var ChallengeTimeout = 1000 * 60  // milliseconds


// makeKey creates a key for a strongly consistent model.
func makeKey(ctx appengine.Context, user_identity, kind string) *datastore.Key {
	parent := datastore.NewKey(ctx, "U2F", "Registration", 0, nil)
	return datastore.NewKey(ctx, kind, user_identity, 0, parent)
}


// NewChallenge creates a new U2F challenge and stores it in the datastore.
//
// Encode the response with e.g.
// 	 json.NewEncoder(w).Encode(req)
//
func NewChallenge(ctx appengine.Context, user_identity string) (*u2f.RegisterRequest, error) {
	// Generate a challenge
	c, err := u2f.NewChallenge(AppID, TrustedFacets); if err != nil {
		// log.Printf("u2f.NewChallenge error: %v", err)
		// http.Error(w, "error", http.StatusInternalServerError)
		return nil, errors.New(fmt.Sprintf("u2f.NewChallenge error: %v", err))
	}

	// Save challenge to database.
	ckey := makeKey(ctx, user_identity, "Challenge")
	if _, err := datastore.Put(ctx, ckey, c); err != nil {
		return nil, errors.New(fmt.Sprintf("datastore.Put error: %v", err))
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
func StoreResponse(ctx appengine.Context, user_identity string, resp u2f.RegisterResponse) error {
	// Load the most recent challenge.
	ckey := makeKey(ctx, user_identity, "Challenge")
  challenge := new(u2f.Challenge)
  if err := datastore.Get(ctx, ckey, challenge); err != nil {
    return errors.New(fmt.Sprintf("datastore.Get error: %v", err))
  }

	if challenge == nil {
		return errors.New(
			fmt.Sprintf("No challenge found for user %v", user_identity))
	}

	// Register the challenge & response
	reg, err := u2f.Register(resp, *challenge, nil)
	if err != nil {
		return errors.New(fmt.Sprintf("u2f.Register error: %v", err))
	}
	buf, err := reg.MarshalBinary()
	if err != nil {
		return errors.New(fmt.Sprintf("reg.MarshalBinary error: %v", err))
	}

	// Save the registration in the datastore
	regi := new(Registration)
	regi.user_identity = user_identity
	regi.counter = 0
	regi.registration = buf
	k := makeKey(ctx, user_identity, "Registration")
  if _, err := datastore.Put(ctx, k, reg); err != nil {
    return errors.New(fmt.Sprintf("datastore.Put error: %v", err))
	}

	return nil
}
