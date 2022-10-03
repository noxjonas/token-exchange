package cognito

import (
	"testing"
)

func TestToOptions(t *testing.T) {
	// must be clear after toOptions
	config.ClientSecret = "an-important-secret"

	args := []string{"https://tx.noxide.xyz", "6eqvuosnu1tuieu5rgdbjg7hq7"}
	err := toOptions(args)
	if err != nil {
		t.Errorf("toOptions: %s", err)
	}

	if config.Domain == "" || config.ClientId == "" || config.LoginUrl == "" {
		t.Error("toOptions: config fields [Domain, ClientId, LoginUrl] must exist after toOptions")
	}
	if config.ClientSecret != "" {
		t.Error("toOptions: config.ClientSecret should be cleared if not provided")
	}

	// must be clear after toOptions
	config.Session.RefreshToken = "an-existing-session-token"

	args = []string{"https://tx.noxide.xyz", "6eqvuosnu1tuieu5rgdbjg7hq7", "secret"}
	err = toOptions(args)
	if err != nil {
		t.Errorf("toOptions: %s", err)
	}

	if config.Session.RefreshToken != "" {
		t.Error("toOptions: stored refresh token must be cleared after new options are read")
	}

	args = []string{"tx.noxide.xyz", "", ""}
	err = toOptions(args)
	if err == nil {
		t.Error("toOptions: domain without protocol (i.e. https://) should throw an error")
	}
}
