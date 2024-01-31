package main

import "net/http"

// Handler fork GET /v1/healthcheck.
// Responds with info about the application, including version and the environment it is running in.
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
		app.serverErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}
}
