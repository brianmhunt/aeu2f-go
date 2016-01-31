package demo


import (
  "log"
  // "fmt"
  "net/http"
  "encoding/json"
  "io/ioutil"

  "appengine"
  // "appengine"
  // "appengine/datastore"
  // "github.com/brianmhunt/aeu2f"

  "github.com/tstranex/u2f"
)

// u2f.Challenge & u2f.Registration
//
// type Challenge struct {
// 	Challenge     []byte
// 	Timestamp     time.Time
// 	AppID         string
// 	TrustedFacets []string
// }
//
// type Registration struct {
// Raw serialized registration data as received from the token.
// 	Raw []byte
//
// 	KeyHandle []byte
// 	PubKey    ecdsa.PublicKey
//
// 	// AttestationCert can be nil for Authenticate requests.
// 	AttestationCert *x509.Certificate
// }
//

type Registered struct {
  Registration []byte
  Counter      int
}


func registerRequest(w http.ResponseWriter, r *http.Request) {
  ctx := appengine.NewContext(r)
  var appID string = appengine.AppID(ctx)
  var trustedFacets = []string{appID}

	c, err := u2f.NewChallenge(appID, trustedFacets)
	if err != nil {
		log.Printf("u2f.NewChallenge error: %v", err)
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}
  // save challenge c

	req := c.RegisterRequest()

	log.Printf("registerRequest: %+v", req)
	json.NewEncoder(w).Encode(req)
}


func indexHandler(w http.ResponseWriter, r *http.Request) {
  b, err := ioutil.ReadFile("index.html")
  if err != nil {
      panic(err)
  }

  w.Write([]byte(b))
}


func init() {
    http.HandleFunc("/", indexHandler)
}
