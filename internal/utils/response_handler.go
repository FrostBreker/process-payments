package utils

import "github.com/gin-gonic/gin"

func SendResponse(c *gin.Context, success bool, status int, errMessage string, message string, data interface{}) {

	//Generate Body
	body := gin.H{
		"success": success,
		"status":  status,
		"error":   errMessage, // Renamed the variable to avoid conflict with the built-in error interface
		"message": message,
		"data":    data,
	}

	//Send Response
	c.JSON(status, body)
}
