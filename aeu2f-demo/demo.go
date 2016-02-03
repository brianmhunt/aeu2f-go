package demo


import (
  "log"
  // "fmt"
  "net/http"
  "encoding/json"
  "io/ioutil"

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


// const appID = "http://localhost:8080/"

// var trustedFacets = []string{appID}

// Normally these state variables would be stored in a database.
// For the purposes of the demo, we just store them in memory.
var challenge *u2f.Challenge
var registration []byte
var counter uint32



type Registered struct {
  Registration []byte
  Counter      int
}

// getAppID returns the U2F application ID, which must be the HTTPS server
// address, e.g. `https://40ccd7b5.ngrok.com`.  There can be no trailing '/'.
// The application must be over HTTPS or the U2F will fail with {errorCode: 2}.
//
// See details of error codes at:
// https://developers.yubico.com/U2F/Libraries/Client_error_codes.html
//
// A simple HTTP -> HTTPS reverse proxy is ngrok.
//
// In production it would defeat a proper security restriction to get the AppID
// from the Referer header, but it is a workaround for dev_appserver.py not
// serving HTTPS.
func getAppID(r *http.Request) string {
  var referer = r.Header["Referer"][0]
  return referer[:len(referer) - 1]
}


func registerRequest(w http.ResponseWriter, r *http.Request) {
  // ctx := appengine.NewContext(r)
  // var appID string = appengine.AppID(ctx)

  var appID = getAppID(r)
  var trustedFacets = []string{appID}

	c, err := u2f.NewChallenge(appID, trustedFacets)
	if err != nil {
		log.Printf("u2f.NewChallenge error: %v", err)
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}
  // save challenge c
  challenge = c

	req := c.RegisterRequest()

	log.Printf("registerRequest: %+v", req)
	json.NewEncoder(w).Encode(req)
}

func registerResponse(w http.ResponseWriter, r *http.Request) {
	var regResp u2f.RegisterResponse
	if err := json.NewDecoder(r.Body).Decode(&regResp); err != nil {
		http.Error(w, "invalid response: "+err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("registerResponse: %+v", regResp)

	if challenge == nil {
		http.Error(w, "challenge not found", http.StatusBadRequest)
		return
	}

	reg, err := u2f.Register(regResp, *challenge, nil)
	if err != nil {
		log.Printf("u2f.Register error: %v", err)
		http.Error(w, "error verifying response", http.StatusInternalServerError)
		return
	}
	buf, err := reg.MarshalBinary()
	if err != nil {
		log.Printf("reg.MarshalBinary error: %v", err)
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}
	registration = buf
	counter = 0

	log.Printf("Registration success: %+v", registration)
	w.Write([]byte("success"))
}

func signRequest(w http.ResponseWriter, r *http.Request) {
  var appID = getAppID(r)
  var trustedFacets = []string{appID}

	if registration == nil {
		http.Error(w, "registration missing", http.StatusBadRequest)
		return
	}

	c, err := u2f.NewChallenge(appID, trustedFacets)
	if err != nil {
		log.Printf("u2f.NewChallenge error: %v", err)
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}
	challenge = c

	var reg u2f.Registration
	if err := reg.UnmarshalBinary(registration); err != nil {
		log.Printf("reg.UnmarshalBinary error: %v", err)
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}

	req := c.SignRequest(reg)
	log.Printf("signRequest: %+v", req)
	json.NewEncoder(w).Encode(req)
}

func signResponse(w http.ResponseWriter, r *http.Request) {
	var signResp u2f.SignResponse
	if err := json.NewDecoder(r.Body).Decode(&signResp); err != nil {
		http.Error(w, "invalid response: "+err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("signResponse: %+v", signResp)

	if challenge == nil {
		http.Error(w, "challenge missing", http.StatusBadRequest)
		return
	}
	if registration == nil {
		http.Error(w, "registration missing", http.StatusBadRequest)
		return
	}

	var reg u2f.Registration
	if err := reg.UnmarshalBinary(registration); err != nil {
		log.Printf("reg.UnmarshalBinary error: %v", err)
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}

	newCounter, err := reg.Authenticate(signResp, *challenge, counter)
	if err != nil {
		log.Printf("VerifySignResponse error: %v", err)
		http.Error(w, "error verifying response", http.StatusInternalServerError)
		return
	}
	log.Printf("newCounter: %d", newCounter)
	counter = newCounter

	w.Write([]byte("success"))
}



func fileHandler(w http.ResponseWriter, r *http.Request) {
  var path string
  if r.URL.Path == "/" {
    path = "index.html"
  } else {
    path = r.URL.Path[1:]
  }
  b, err := ioutil.ReadFile(path)
  if err != nil {
      panic(err)
  }

  w.Write([]byte(b))
}


func init() {
    http.HandleFunc("/", fileHandler)
  	http.HandleFunc("/registerRequest", registerRequest)
  	http.HandleFunc("/registerResponse", registerResponse)
  	http.HandleFunc("/signRequest", signRequest)
  	http.HandleFunc("/signResponse", signResponse)
    // TODO: Delete.
}
