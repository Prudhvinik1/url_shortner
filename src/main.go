package main

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

type urls struct {
	ID          string `json:"id"`
	ShortCode   string `json:"short_code"`
	OriginalURL string `json:"original_url"`
	IsAlias     bool   `json:"isAlias"`
	TTL         int64  `json:"ttl"`
}

type Handler struct {
	DB *sql.DB
}

func (h *Handler) getLongUrl_service(c *gin.Context) {
	short_code := c.Param("short_code")
	long_url, err := getLongUrl(h.DB, short_code)
	if err != nil {
		if err == sql.ErrNoRows {
			c.IndentedJSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, long_url)
}

func (h *Handler) postLongUrl(c *gin.Context) {
	var newUrl urls

	//Call BindJson to bind the recieved JSON
	if err := c.BindJSON(&newUrl); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	const maxAttempts = 5
	var candidate_shortcode string
	var genErr error

	for attempts := 0; attempts < maxAttempts; attempts++ {
		candidate_shortcode, genErr = generateShortCode()
		if genErr == nil {
			break
		}
		fmt.Println("Generate ShortCode Error: ", genErr)
	}

	if genErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate shortcode"})
		return
	}

	if err := insertUrl(h.DB, candidate_shortcode, newUrl.OriginalURL, newUrl.IsAlias, newUrl.TTL, 33); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusAccepted, candidate_shortcode)
}

func main() {
	gin.SetMode(gin.DebugMode)
	_ = godotenv.Load()

	db := initDB()
	defer db.Close() // Scheduled to run when main() returns

	h := &Handler{DB: db}

	router := gin.Default()
	router.GET("/:short_code", h.getLongUrl_service)
	router.POST("/urls", h.postLongUrl)
	router.Run("localhost:8080")
}
