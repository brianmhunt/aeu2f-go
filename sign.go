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


func signChallengeRequest(c u2f.Challenge, regi Registration) (*u2f.SignRequest, error) {
	var reg u2f.Registration
	buf := regi.U2FRegistrationBytes

	if err := reg.UnmarshalBinary(buf); err != nil {
		return nil, fmt.Errorf("reg.UnmarshalBinary %v", err)
	}

	return c.SignRequest(reg), nil
}


// NewSignChallenge returns a challenge for the U2F device.
//
func NewSignChallenge(ctx appengine.Context, userIdentity string) ([]*u2f.SignRequest, error) {

	// Create challenge
	c, err := u2f.NewChallenge(AppID, TrustedFacets); if err != nil {
		return nil, fmt.Errorf("u2f.NewChallenge error: %v", err)
	}

	// Load Registrations
	pkey := MakeParentKey(ctx)
	q := datastore.NewQuery("Registration").
		Ancestor(pkey).
		Filter("UserIdentity =", userIdentity)

	var reqs = []*u2f.SignRequest{}

	// Retrieve & convert registrations into  array of u2f challenges
	for t := q.Run(ctx) ; ; {
		var regi Registration
		_, err = t.Next(&regi)
		if err == datastore.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("datastore query error: %+v", err)
		}

		signr, err := signChallengeRequest(*c, regi)
		if err != nil {
			return nil, fmt.Errorf("Signing error: %+v", err)
		}

		reqs = append(reqs, signr)
	}

	// Save challenge to database.
	ckey := makeKey(ctx, userIdentity, "SignChallenge")
	if _, err := datastore.Put(ctx, ckey, c); err != nil {
		return nil, fmt.Errorf("datastore.Put error: %v", err)
	}

	// Return challenge
	log.Printf("🖋  New Sign Challenges: %+v [%+v]", reqs, ckey)
	return reqs, nil
}


// Sign verifies or rejects a U2F response.
func Sign(ctx appengine.Context, userIdentity string, resp u2f.RegisterResponse) error {
  return nil
}
