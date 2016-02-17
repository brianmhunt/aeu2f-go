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


// loadRegistrations returns a slice of the registrations for a given user
// identity.
func loadRegistrations(ctx appengine.Context, userIdentity string) ([]*datastore.Key, []*Registration, error) {
	regis := []*Registration{}
	// Load Registrations
	pkey := MakeParentKey(ctx)
	q := datastore.NewQuery("Registration").
		Ancestor(pkey).
		Filter("UserIdentity =", userIdentity)

	// Retrieve & save registrations into array of challenges
	keys, err := q.GetAll(ctx, &regis)
	if err != nil {
		return nil, nil, fmt.Errorf("datastore GetAll error: %+v", err)
	}

	return keys, regis, nil
}


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

	_, regis, err := loadRegistrations(ctx, userIdentity)
	if err != nil {
		return nil, fmt.Errorf("loadRegistrations %+v", err)
	}

	var reqs = []*u2f.SignRequest{}
	for _, regi := range regis {
		signr, err := signChallengeRequest(*c, *regi)
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


// --- testSignChallenge ---
func testSignChallenge(challenge u2f.Challenge, regi Registration, signResp u2f.SignResponse) error {
	var reg u2f.Registration
	if err := reg.UnmarshalBinary(regi.U2FRegistrationBytes); err != nil {
		return fmt.Errorf("reg.UnmarshalBinary error: %v", err)
	}

	// The AppEngine datastore does not accept uint types, see:
	// https://github.com/golang/appengine/blob/master/datastore/save.go#L148
	// So we cast int64 to uint32 when coming from the datastore, and back.
	newCounter, err := reg.Authenticate(signResp, challenge, uint32(regi.Counter))
	if err != nil {
		return fmt.Errorf("VerifySignResponse error: %v", err)
	}

	// Update the counter for the next auth.
	regi.Counter = int64(newCounter)
	return nil
}


// Sign verifies or rejects a U2F response.
func Sign(ctx appengine.Context, userIdentity string, signResp u2f.SignResponse) error {
	ckey := makeKey(ctx, userIdentity, "SignChallenge")
	var challenge u2f.Challenge

	// Load the Challenge for this user
	if err := datastore.Get(ctx, ckey, &challenge); err != nil {
		return fmt.Errorf("datastore.Get error: %v", err)
	}

	// Load the Registrations
	keys, regis, err := loadRegistrations(ctx, userIdentity)
	if err != nil {
		return fmt.Errorf("loadRegistrations error %+v", err)
	}

	// Check each Registration
	for idx, regi := range regis {
		if err := testSignChallenge(challenge, *regi, signResp); err != nil {
			return fmt.Errorf("Sign error: %v", err)
		} else {
			// Update the counter for the regi.
			if _, err := datastore.Put(ctx, ckey, keys[idx]); err != nil {
				return fmt.Errorf("datastore.Put error: %v", err)
			}

			// Success -- A U2F response to a sign challenge succeeded.
			return nil
		}
		return fmt.Errorf("Challenge failed for known registrations.")
	}

	return nil
}
