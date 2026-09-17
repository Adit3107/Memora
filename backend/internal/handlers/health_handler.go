package handlers

import "github.com/gin-gonic/gin"

//Handler = function that receives HTTP request and sends HTTP response

func HealthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "ok",
	})
}

//gin.H is a shortcut for map[string]interface{} in the Gin framework. It is used to create a JSON response easily. In this case, it creates a JSON object with a single key-value pair: "status": "ok". This indicates that the health check endpoint is functioning correctly and the server is healthy.

//gin.Context is a struct that holds the request and response objects, path parameters, and other information about the HTTP request. It is used to handle HTTP requests and responses in the Gin web framework. In this case, it is used to send a JSON response with a status of "ok" when the health check endpoint is called.
