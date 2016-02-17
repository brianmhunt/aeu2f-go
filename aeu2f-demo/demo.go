package demo


import (
  "log"
  "fmt"
  "net/http"
  "encoding/json"
  "io/ioutil"

  "appengine"
  "appengine/datastore"
  "github.com/brianmhunt/aeu2f-go"

  "github.com/tstranex/u2f"
)

// Handle /register/USERNAME and /auth/USERNAME, respectively.
const registerURLPrefix = "/register/"
const authURLPrefix = "/auth/"
const listURLPrefix = "/list/"


// HTTP request wrappers
// https://gist.github.com/tristanwietsma/8444cf3cb5a1ac496203


// GetAppID returns a U2F application ID, which must be the HTTPS server
// address, e.g. `https://40ccd7b5.ngrok.com`.  There can be no trailing '/'.
//
// Also, The application must be over HTTPS or the U2F will fail with
// {errorCode: 2}.
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

// func signRequest(w http.ResponseWriter, r *http.Request) {
//   var appID = getAppID(r)
//   var trustedFacets = []string{appID}
//
// 	if registration == nil {
// 		http.Error(w, "registration missing", http.StatusBadRequest)
// 		return
// 	}
//
// 	c, err := u2f.NewChallenge(appID, trustedFacets)
// 	if err != nil {
// 		log.Printf("u2f.NewChallenge error: %v", err)
// 		http.Error(w, "error", http.StatusInternalServerError)
// 		return
// 	}
// 	challenge = c
//
// 	var reg u2f.Registration
// 	if err := reg.UnmarshalBinary(registration); err != nil {
// 		log.Printf("reg.UnmarshalBinary error: %v", err)
// 		http.Error(w, "error", http.StatusInternalServerError)
// 		return
// 	}
//
// 	req := c.SignRequest(reg)
// 	log.Printf("signRequest: %+v", req)
// 	json.NewEncoder(w).Encode(req)
// }
//
// func signResponse(w http.ResponseWriter, r *http.Request) {
// 	var signResp u2f.SignResponse
// 	if err := json.NewDecoder(r.Body).Decode(&signResp); err != nil {
// 		http.Error(w, "invalid response: "+err.Error(), http.StatusBadRequest)
// 		return
// 	}
//
// 	log.Printf("signResponse: %+v", signResp)
//
// 	if challenge == nil {
// 		http.Error(w, "challenge missing", http.StatusBadRequest)
// 		return
// 	}
// 	if registration == nil {
// 		http.Error(w, "registration missing", http.StatusBadRequest)
// 		return
// 	}
//
// 	var reg u2f.Registration
// 	if err := reg.UnmarshalBinary(registration); err != nil {
// 		log.Printf("reg.UnmarshalBinary error: %v", err)
// 		http.Error(w, "error", http.StatusInternalServerError)
// 		return
// 	}
//
// 	newCounter, err := reg.Authenticate(signResp, *challenge, counter)
// 	if err != nil {
// 		log.Printf("VerifySignResponse error: %v", err)
// 		http.Error(w, "error verifying response", http.StatusInternalServerError)
// 		return
// 	}
// 	log.Printf("newCounter: %d", newCounter)
// 	counter = newCounter
//
// 	w.Write([]byte("success"))
// }
//  ^^^^^^^^^^^^^


// --- fileHandler ---
//
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

// --- setupUserContext ---
//
func setupUserContext(r *http.Request, prefix string) (appengine.Context, string) {
  // Set up the aeu2f variables.
  aeu2f.AppID = getAppID(r)
  aeu2f.TrustedFacets = []string{aeu2f.AppID}

  // Create an AppEngine context
  ctx := appengine.NewContext(r)

  // Get the user identity
  userIdentity := r.URL.Path[len(prefix):]
  return ctx, userIdentity
}


// --- createRegistrationChallenge ---
//
func createRegistrationChallenge(ctx appengine.Context, userIdentity string) (interface{}, error) {
  req, err := aeu2f.NewRegistrationChallenge(ctx, userIdentity)
  if err != nil {
    return nil, fmt.Errorf("Registration Challenge error: %v", err)
  }

	log.Printf("Created Registration Challenge: %+v", req)
  return req, nil
}


// --- testRegistrationResponse ---
//
func testRegistrationResponse(ctx appengine.Context, userIdentity string, regResp u2f.RegisterResponse) (interface{}, error) {
	if err := aeu2f.StoreResponse(ctx, userIdentity, regResp); err != nil {
    return nil, fmt.Errorf("Registration error: %v", err)
  }
  return "success", nil
}


// --- registerHandler ---
//
func registerHandler(w http.ResponseWriter, r *http.Request) {
  ctx, userIdentity := setupUserContext(r, registerURLPrefix)
  if userIdentity == "" {
    http.Error(w, "User identity not provided", http.StatusBadRequest)
  }

  var err error
  var ret interface{}

  switch r.Method {

  case "GET":
    ret, err = createRegistrationChallenge(ctx, userIdentity)

  case "POST":
  	var regResp u2f.RegisterResponse
  	if err := json.NewDecoder(r.Body).Decode(&regResp); err != nil {
  		http.Error(w, "invalid response: "+err.Error(), http.StatusBadRequest)
  		return
  	}

  	log.Printf("Registration Response: %+v", regResp)

    ret, err = testRegistrationResponse(ctx, userIdentity, regResp)
  default:
    http.Error(w, "Method not supported.", http.StatusBadRequest)
  }

  if err != nil {
    http.Error(w, "Error: %v" + err.Error(), http.StatusBadRequest)
  }

  json.NewEncoder(w).Encode(ret)
}

// --- createAuthChallenge ---
//
func createAuthChallenge(ctx appengine.Context, userIdentity string) (interface{}, error) {
  reqs, err := aeu2f.NewSignChallenge(ctx, userIdentity)
  if err != nil {
    return nil, fmt.Errorf("Auth Challenge error: %v", err)
  }

	log.Printf("Created Auth Challenge(s): %+v", reqs)
  return reqs, nil
}

// --- testAuthResponse ---
func testAuthResponse(ctx appengine.Context, userIdentity string, signResp u2f.SignResponse) (interface{}, error) {

  if err := aeu2f.Sign(ctx, userIdentity, signResp); err != nil {
    return nil, fmt.Errorf("Sign failure: %v", err)
  }

  return "success", nil
}


// --- authHandler ---
//
func authHandler(w http.ResponseWriter, r *http.Request) {
  ctx, userIdentity := setupUserContext(r, authURLPrefix)
  if userIdentity == "" {
    http.Error(w, "User identity not provided", http.StatusBadRequest)
  }
  var err error
  var ret interface{}

  switch r.Method {

  case "GET":
    ret, err = createAuthChallenge(ctx, userIdentity)

  case "POST":
  	var signResp u2f.SignResponse
  	if err := json.NewDecoder(r.Body).Decode(&signResp); err != nil {
  		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
  		return
  	}

  	log.Printf("Auth Response: %+v", signResp)

    ret, err = testAuthResponse(ctx, userIdentity, signResp)
  default:
    http.Error(w, "Method not supported.", http.StatusBadRequest)
  }

  if err != nil {
    log.Printf("testAuthResponse error: %+v", err)
    http.Error(w, "Error: %v" + err.Error(), http.StatusBadRequest)
  }

  json.NewEncoder(w).Encode(ret)
}

// --- listHandler ---
// Return a list of the keys for the given user.
func listHandler(w http.ResponseWriter, r *http.Request) {
  ctx, userIdentity := setupUserContext(r, listURLPrefix)
  if userIdentity == "" {
    http.Error(w, "User identity not provided", http.StatusBadRequest)
  }

  reqs := []aeu2f.Registration{}
  q := datastore.NewQuery("Registration").
    Ancestor(aeu2f.MakeParentKey(ctx)).
    Filter("UserIdentity =", userIdentity)

	for t := q.Run(ctx) ; ; {
		var regi aeu2f.Registration
		if _, err := t.Next(&regi); err == datastore.Done {
			break
		} else if err != nil {
  		http.Error(w, "datastore error: "+err.Error(), http.StatusBadRequest)
			return
		}

		reqs = append(reqs, regi)
	}

  json.NewEncoder(w).Encode(reqs)
}

// --- init ---
//
func init() {
    http.HandleFunc("/", fileHandler)

    http.HandleFunc(registerURLPrefix, registerHandler)
    http.HandleFunc(authURLPrefix, authHandler)
    http.HandleFunc(listURLPrefix, listHandler)
    // TODO: Delete.
}
