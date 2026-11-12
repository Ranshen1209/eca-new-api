package controller

import (
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
)

// Check local account endpoints, including continuations of existing flows.
func requireLocalAccountManagement(c *gin.Context, userID int) bool {
	user, err := model.GetUserById(userID, false)
	if err != nil {
		writeSecurityOperationError(c, err)
		return false
	}
	if user.IdentityManaged() {
		writeSecurityOperationError(c, model.ErrIdentityManaged)
		return false
	}
	return true
}
