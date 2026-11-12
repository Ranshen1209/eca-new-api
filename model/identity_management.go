package model

import "errors"

var ErrIdentityManaged = errors.New("This account is managed by your identity provider. Manage your sign-in details and security in the identity center.")

// The persisted binding owns this decision, including OIDC administrators.
// Unbound operational accounts retain their local credentials.
func (user *User) IdentityManaged() bool { return user.OidcId != "" }
