package main

import (
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/firestore"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/iterator"
)

const (
	queryLimit = 32
)

var (
	projectID        = ""
	typeToCollection = map[string]string{
		"file":    "files",
		"command": "commands",
		"ip":      "ips",
	}
)

func main() {
	projectID = os.Getenv("GOOGLE_CLOUD_PROJECT")

	router := gin.Default()
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:8080", "https://ossf-malware-analysis.storage.googleapis.com"}
	router.Use(cors.New(config))

	router.POST("/query", queryHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("defaulting to port %s", port)
	}

	// Start HTTP server.
	log.Printf("listening on port %s", port)
	router.Run()
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

type Request struct {
	Type   string `json:"type"`
	Search string `json:"search"`
	Cursor string `json:"cursor"`
}

type Response struct {
	Packages []map[string]interface{} `json:"packages"`
	Next     string                   `json:"next"`
}

func queryHandler(c *gin.Context) {
	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collection := typeToCollection[req.Type]
	if collection == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid search type"})
		return
	}

	if req.Search == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "search value should not be empty"})
		return
	}

	fs, err := firestore.NewClient(c, projectID)
	if err != nil {
		log.Printf("Failed to create firestore client: %v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	defer fs.Close()

	q := fs.Collection(collection).Select("Package").
		Where("Indexes", "array-contains", req.Search).
		OrderBy("__name__", firestore.Asc).
		Limit(queryLimit + 1) // Check if there are more by fetching one more.

	if req.Cursor != "" {
		q = q.StartAt(req.Cursor)
	}

	it := q.Documents(c)
	defer it.Stop()

	resp := Response{}
	for {
		doc, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("Failed to iterate: %v", err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		if len(resp.Packages) < queryLimit {
			resp.Packages = append(resp.Packages, doc.Data())
		} else {
			// If we get here, we're at the 1 extra document we fetched,
			// so this is the cursor for the next page.
			resp.Next = doc.Ref.ID
		}
	}

	c.JSON(http.StatusOK, &resp)
}
