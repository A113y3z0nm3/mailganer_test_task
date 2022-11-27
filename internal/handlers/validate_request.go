package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// used to help extract validation errors
type invalidArgument struct {
	Field string      `json:"field"`
	Value interface{} `json:"value"`
	Tag   string      `json:"tag"`
	Param string      `json:"param"`
}

// bindData is helper function, returns false if data is not bound
func bindData(c *gin.Context, req interface{}) bool {
	if c.ContentType() != "application/json" {

		c.JSON(http.StatusUnsupportedMediaType, gin.H{
			"error": "Endpoint only accepts Content-Type application/json",
		})

		return false
	}
	// Bind incoming json to struct and check for validation errors
	if err := c.ShouldBind(req); err != nil {
		log.Printf("Error binding data: %+v\n", err)

		if errs, ok := err.(validator.ValidationErrors); ok {
			// could probably extract this, it is also in middleware_auth_user
			var invalidArgs []invalidArgument

			for _, err := range errs {
				invalidArgs = append(invalidArgs, invalidArgument{
					err.Field(),
					err.Value(),
					err.Tag(),
					err.Param(),
				})
			}

			c.JSON(http.StatusBadRequest, gin.H{
				"error":       "Invalid request parameters. See invalidArgs",
				"invalidArgs": invalidArgs,
			})

			return false
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})

		return false
	}

	return true
}
