// Package routes for API endpoints
package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	c4sDeployer "github.com/codescalers/cloud4students/deployer"
	"github.com/codescalers/cloud4students/internal"
	"github.com/codescalers/cloud4students/middlewares"
	"github.com/codescalers/cloud4students/models"
	"github.com/codescalers/cloud4students/streams"
	"github.com/codescalers/cloud4students/validators"
	"github.com/rs/zerolog/log"
	"gopkg.in/validator.v2"
)

const internalServerErrorMsg = "Something went wrong"

// Router struct holds db model and configurations

/*
Either this structure is mis-named, or can be instead merged with the App struct into one.

This is called Router, but I don't see it do any routing, imho this Struct is only holding the actual
handlers to the API endpoints! It then can be called, API, or Handlers for example.

That being said, if that is the case, may be it is can also be merged with the App structure then.
IMHO a simpler design will be like

type App struct{}

// define handlers endpoint
```go

	func (a *App) someHandler(request) {}

	func (a *App) Start() error {
		r := mux.Router()
		r.HandleFunc(/path, a.someHandler)

		server := // create server
		return server.Start()
	}

```
*/
type Router struct {
	config   *internal.Configuration
	db       models.DB
	deployer c4sDeployer.Deployer
}

// NewRouter create new router with db
func NewRouter(config internal.Configuration, db models.DB, redis streams.RedisClient, deployer c4sDeployer.Deployer) (Router, error) {
	// validations
	/*
		I really think this in the wrong place. I think validator registration should be instead defined
		in a package level `init` function

		like
		```go
		func init() {
			// register custom validators here.
		}
		```
	*/
	err := validator.SetValidationFunc("ssh", validators.ValidateSSHKey)
	if err != nil {
		return Router{}, err
	}
	err = validator.SetValidationFunc("password", validators.ValidatePassword)
	if err != nil {
		return Router{}, err
	}
	err = validator.SetValidationFunc("mail", validators.ValidateMail)
	if err != nil {
		return Router{}, err
	}

	return Router{
		&config,
		db,
		deployer,
	}, nil
}

// ErrorMsg holds errors
type ErrorMsg struct {
	Error string `json:"err"`
}

// ResponseMsg holds messages and needed data
type ResponseMsg struct {
	Message string      `json:"msg"`
	Data    interface{} `json:"data,omitempty"`
}

/*
I really HATE that you have to call writeErrResponse and writeMsgResponse IN all handlers. this makes
building a handler (that can fail at many places) extremely annoying and difficult and not readable.

Instead I usually create a "Wrapper" around action function that takes care of this, then my action/handler function
can simply return (result, error)

for example

func MyHandler(request *http.Request)(interface{}, error) {
	if something {
		return nil, error
	}
	// if success return
	return data, nil
}

then on the router I can do
r.HandleFunc(/path, Wrapper(MyHandler))

I implemented this multiple times before (even in zos) for example check this implementation
https://github.com/threefoldtech/zos/blob/main/pkg/provision/mw/action.go#L27

This code is NOT used anymore in ZOS but i still have it just in case

this also gives u full control on what error code to return (or Ok) and also add extra
header if u want.
*/

// writeErrResponse write error messages in api
func writeErrResponse(r *http.Request, w http.ResponseWriter, statusCode int, errStr string) {
	jsonErrRes, _ := json.Marshal(ErrorMsg{Error: errStr})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_, err := w.Write(jsonErrRes)
	if err != nil {
		log.Error().Err(err).Msg("write error response failed")
	}
	middlewares.Requests.WithLabelValues(r.Method, r.RequestURI, fmt.Sprint(statusCode)).Inc()
}

// writeMsgResponse write response messages for api
func writeMsgResponse(r *http.Request, w http.ResponseWriter, message string, data interface{}) {
	contentJSON, err := json.Marshal(ResponseMsg{Message: message, Data: data})
	if err != nil {
		log.Error().Err(err).Send()
		writeErrResponse(r, w, http.StatusInternalServerError, internalServerErrorMsg)
		middlewares.Requests.WithLabelValues(r.Method, r.RequestURI, fmt.Sprint(http.StatusInternalServerError)).Inc()
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(contentJSON)
	if err != nil {
		log.Error().Err(err).Msg("write error response failed")
		writeErrResponse(r, w, http.StatusInternalServerError, internalServerErrorMsg)
		middlewares.Requests.WithLabelValues(r.Method, r.RequestURI, fmt.Sprint(http.StatusInternalServerError)).Inc()
		return
	}

	middlewares.Requests.WithLabelValues(r.Method, r.RequestURI, fmt.Sprint(http.StatusOK)).Inc()
}
