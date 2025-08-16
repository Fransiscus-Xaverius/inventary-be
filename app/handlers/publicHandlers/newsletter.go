package publicHandlers

import (
	"net/http"

	"github.com/everysoft/inventary-be/app/handlers"
	"github.com/everysoft/inventary-be/app/models"
	"github.com/everysoft/inventary-be/db"
	"github.com/gin-gonic/gin"
)

// SubscribeToNewsletter handles new newsletter subscriptions
func SubscribeToNewsletter(c *gin.Context) {
	var newsletter models.Newsletter
	if err := c.ShouldBindJSON(&newsletter); err != nil {
		handlers.SendError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if err := db.InsertNewsletter(&newsletter); err != nil {
		handlers.SendError(c, http.StatusInternalServerError, "Failed to subscribe to newsletter: "+err.Error(), nil)
		return
	}

	handlers.SendSuccess(c, http.StatusCreated, gin.H{"message": "Subscription successful"})
}
