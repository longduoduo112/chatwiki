// Copyright © 2016- 2025 Wuhan Sesame Small Customer Service Network Technology Co., Ltd.

package business

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Completions compatible openai standard api
func Completions(c *gin.Context) {
	c.String(http.StatusNotFound, `The open-source version is not supported !`)
}
