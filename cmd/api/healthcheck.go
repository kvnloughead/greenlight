package main

import (
	"net/http"
)

// healthcheck handles GET requests to the /v1/healthcheck endpoint.
// It responds with the current status of the application and some system info,
// including the environment and version. This endpoint can be used as a service
// health check to monitor the application's status and gather basic system
// information.
//
// Responds with a JSON object in the following format:
//
//	{
//	  "status": "available",
//	  "system_info": {
//				"environment": <app_environment>,
//				"version":     <app_version>,
//	  }
//	}
//
// If the app is unable to construct the response a 500 Internal Server Error
// is sent with no body.
func (app *application) healthcheck(w http.ResponseWriter, r *http.Request) {
	env := envelope{
		"status": "available",
		"system_info": map[string]string{
			"environment": app.config.env,
			"version":     version,
		},
	}

	err := app.writeJSON(w, http.StatusOK, env, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}
