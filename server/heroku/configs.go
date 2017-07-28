package heroku

import (
	"net/http"

	"github.com/remind101/empire"
	"github.com/remind101/empire/server/auth"
)

func (h *Server) GetConfigs(w http.ResponseWriter, r *http.Request) error {
	a, err := h.findApp(r)
	if err != nil {
		return err
	}

	c, err := h.Config(a)
	if err != nil {
		return err
	}

	w.WriteHeader(200)
	return Encode(w, c.Vars)
}

func (h *Server) PatchConfigs(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	var configVars empire.Vars

	if err := Decode(r, &configVars); err != nil {
		return err
	}

	a, err := h.findApp(r)
	if err != nil {
		return err
	}

	m, err := findMessage(r)
	if err != nil {
		return err
	}

	// Update the config
	c, err := h.Set(ctx, empire.SetOpts{
		User:    auth.UserFromContext(ctx),
		App:     a,
		Vars:    configVars,
		Message: m,
	})
	if err != nil {
		return err
	}

	w.WriteHeader(200)
	return Encode(w, c.Vars)
}
