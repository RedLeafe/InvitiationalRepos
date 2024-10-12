// This simply provides tools to create a session for the user as a text file on disk in /tmp/usersession, as well as a function to check if the user credentials exist in active directory

package auth

import (
	"fmt"

	auth "github.com/korylprince/go-ad-auth/v3"
)

// login to active directory
func Login(username, password string) error {
	config := &auth.Config{
		Server: "kerberos.alien.moon.mine",
		Port:   389,
		BaseDN: "OU=Users,dc=alien,dc=moon,dc=mine",
	}

	status, err := auth.Authenticate(config, username, password)
	if err != nil {
		return err
	}

	if !status {
		return fmt.Errorf("authentication failed")
	}

	return nil
}
