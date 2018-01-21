// A learning excercise to understand the use of the Go programming language when building an API
// Started Christmas 2017
// Governed by the license that can be found in the LICENSE file
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/tkanos/gonfig"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/lib/pq"

	"github.com/deezone/rest-api-go/toolbox"
)

type Health struct {
	Refer      string `json:"reference,omitempty"`
	Alloc      uint64 `json:"alloc,omitempty"`
	TotalAlloc uint64 `json:"total-alloc,omitempty"`
	Sys        uint64 `json:"sys,omitempty"`
	NumGC      uint32 `json:"numgc,omitempty"`
}

type Ready struct {
	Ready string `json:"ready,omitempty"`
}

type Version struct {
	Version string     `json:"version,omitempty"`
	ReleaseDate string `json:"release-date,omitempty"`
}

var quotes []toolbox.Quote
var quotesmin []toolbox.QuoteMin
var authors []toolbox.Author
var authorsmin []toolbox.AuthorMin
var err error

// GetAuthors looks up all of the authors.
// GET /authors
// Populates authors slice with all of the author records in the database and returns JSON formatted listing.
// @todo: exclude author information in quotes list with each author
func GetAuthors(w http.ResponseWriter, r *http.Request) {
	count := 0
	authors = []toolbox.Author{}
	toolbox.Db.Find(&authors).Count(&count)
	if count == 0 {
		toolbox.RespondWithError(w, http.StatusOK, "Author records not found.")
		return
	}

	// Lookup author quotes
	// @todo: ISSUE-16 - create parameter to trigger author lookup rather than being the default response
	// @todo: ISSUE-17 - create parameter to include deleted quotes in response
	quotesmin := []toolbox.QuoteMin{}
	for index, author := range authors {
		toolbox.Db.Raw("SELECT * FROM quotes WHERE author_id = ? AND deleted_at IS NULL", author.ID).Scan(&quotesmin)
		authors[index].Quotes = quotesmin
	}

	toolbox.RespondWithJSON(w, http.StatusOK, authors)
}

// GetAuthor looks up a specific author by ID.
// GET /author
// Looks up a author in the database by ID and returns results JSON format.
func GetAuthor(w http.ResponseWriter, r *http.Request) {
	var author toolbox.Author

	params := mux.Vars(r)
	authorID, err := strconv.Atoi(params["id"])
	if (err != nil) {
		toolbox.RespondWithError(w, http.StatusBadRequest, "Invalid author ID")
		return
	}

	// Check that author ID is valid
	if (toolbox.Db.First(&author, authorID).RecordNotFound()) {
		message := []string{}
		message = append(message, "Author ID: ", strconv.Itoa(int(authorID)), " not found.")
		toolbox.RespondWithError(w, http.StatusBadRequest, strings.Join(message, ""))
		return
	}

	// Lookup author quotes
	// @todo: ISSUE-16 - create parameter to trigger this lookup rather than being the default
	// @todo: ISSUE-17 - create parameter to include deleted quotes in response
	quotesmin := []toolbox.QuoteMin{}
	toolbox.Db.Raw("SELECT * FROM quotes WHERE author_id = ? AND deleted_at IS NULL", author.ID).Scan(&quotesmin)
	author.Quotes = quotesmin

	toolbox.RespondWithJSON(w, http.StatusOK, author)
}

// CreateAuthor creates a new author.
// POST /author
// Returns newly created author ID.
func CreateAuthor(w http.ResponseWriter, r *http.Request) {

	var author toolbox.Author
	_ = json.NewDecoder(r.Body).Decode(&author)

	// Create new record
	if err := toolbox.Db.Create(&author).Error; err != nil {
		toolbox.RespondWithError(w, http.StatusBadRequest, "Error creatng author record.")
		return
	}

	message := []string{}
	message = append(message, "Author ID: ", strconv.Itoa(int(author.ID)), " created.")
	toolbox.RespondWithJSON(w, http.StatusCreated, map[string]string{"status": strings.Join(message, "")})
}

// Delete Author deletes an author by author ID.
// DELETE /author/{id}
// Returns a status message that includes the ID of the author record deleted.
func DeleteAuthor(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	authorID, err := strconv.Atoi(params["id"])
	if (err != nil) {
		toolbox.RespondWithError(w, http.StatusBadRequest, "Invalid author ID")
		return
	}

	message := []string{}
	var author toolbox.Author
	if (toolbox.Db.First(&author, authorID).RecordNotFound()) {
		message = append(message, "Author ID: ", strconv.Itoa(authorID), " not found.")
		toolbox.RespondWithError(w, http.StatusBadRequest, strings.Join(message, ""))
		return
	}
	toolbox.Db.Delete(&author)

	// @todo: ISSUE-18 - delete quotes attributed to deleted author

	message = append(message, "Author ID: ", strconv.Itoa(authorID), " deleted.")
	toolbox.RespondWithJSON(w, http.StatusOK, map[string]string{"status": strings.Join(message, "")})
}

// GetQuotes looks up all of the quotes.
// GET /quotes
// Returns all of the quotes in JSON format.
func GetQuotes(w http.ResponseWriter, r *http.Request) {
	count := 0
	quotes = []toolbox.Quote{}
	toolbox.Db.Find(&quotes).Count(&count)
	if count == 0 {
		toolbox.RespondWithError(w, http.StatusOK, "Quote records not found.")
		return
	}

	// Lookup quote author
	// @todo: ISSUE-16 - create parameter to trigger author lookup rather than being the default response
	authormin := toolbox.AuthorMin{}
	for index, quote := range quotes {
		toolbox.Db.Raw("SELECT * FROM authors WHERE id = ? AND deleted_at IS NULL", quote.AuthorID).Scan(&authormin)
		quotes[index].Author = authormin
	}

	toolbox.RespondWithJSON(w, http.StatusOK, quotes)
}

// GetQuote looks up a specific quote by ID.
// GET /quote/{id}
// Returns a quote in the JSON format provided the target ID is valid.
func GetQuote(w http.ResponseWriter, r *http.Request) {
	var quote toolbox.Quote

	params := mux.Vars(r)
	quoteID, err := strconv.Atoi(params["id"])
	if (err != nil) {
		toolbox.RespondWithError(w, http.StatusBadRequest, "Invalid quote ID")
		return
	}

	// Check that quote ID is valid
	if (toolbox.Db.First(&quote, quoteID).RecordNotFound()) {
		message := []string{}
		message = append(message, "Quote ID: ", strconv.Itoa(int(quoteID)), " not found.")
		toolbox.RespondWithError(w, http.StatusBadRequest, strings.Join(message, ""))
		return
	}

	// Lookup quote author
	// @todo: ISSUE-16 - create parameter to trigger author lookup rather than the default
	authormin := toolbox.AuthorMin{}
	toolbox.Db.Raw("SELECT * FROM authors WHERE id = ? AND deleted_at IS NULL", quote.AuthorID).Scan(&authormin)
	quote.Author = authormin

	toolbox.RespondWithJSON(w, http.StatusOK, quote)
}

// CreateQuote creates a new quote. Validates that the author ID exists.
// POST /quote
// Returns the ID of new quote as a part of the "status" response message.
func CreateQuote(w http.ResponseWriter, r *http.Request) {

	message := []string{}
	var quote toolbox.Quote
	_ = json.NewDecoder(r.Body).Decode(&quote)

	// Validate that the author ID exists
	var author toolbox.Author
	if (toolbox.Db.First(&author, quote.AuthorID).RecordNotFound()) {
		message = append(message, "Invalid author, authorid: ", strconv.Itoa(int(quote.AuthorID)), " not found.")
		toolbox.RespondWithError(w, http.StatusBadRequest, strings.Join(message, ""))
		return
	}

	toolbox.Db.Create(&quote)
	message = append(message, "Quote ID: ", strconv.Itoa(int(quote.ID)), " created for authorID: ",
		strconv.Itoa(int(quote.AuthorID)), ".")
	toolbox.RespondWithJSON(w, http.StatusCreated, map[string]string{"status": strings.Join(message, "")})
}

// DeleteQuote deletes a quote by quote ID.
// DELETE /quote/{id}
// Returns.
func DeleteQuote(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	quoteID, err := strconv.Atoi(params["id"])
	if (err != nil) {
		toolbox.RespondWithError(w, http.StatusBadRequest, "Invalid quote ID")
		return
	}

	message := []string{}
	var quote toolbox.Quote
	if (toolbox.Db.First(&quote, quoteID).RecordNotFound()) {
		message = append(message, "Quote ID: ", strconv.Itoa(quoteID), " not found.")
		toolbox.RespondWithError(w, http.StatusBadRequest, strings.Join(message, ""))
		return
	}
	toolbox.Db.Delete(&quote)

	// @todo: remove author ID from quotes that reference the deleted author

	message = append(message, "Quote ID: ", strconv.Itoa(quoteID), " deleted.")
	toolbox.RespondWithJSON(w, http.StatusOK, map[string]string{"status": strings.Join(message, "")})
}

// GetHealth looks up the health of the application.
// GET /health
// Returns all of the health status of all the components of the application.
func GetHealth(w http.ResponseWriter, r *http.Request) {
	var m runtime.MemStats
	var data Health

	runtime.ReadMemStats(&m)

	data.Refer = "https://golang.org/pkg/runtime/#MemStats"
	data.Alloc = m.Alloc / 1024
	data.TotalAlloc = m.TotalAlloc / 1024
	data.Sys = m.Sys / 1024
	data.NumGC = m.NumGC

	toolbox.RespondWithJSON(w, http.StatusOK, data)
}

// GetVersion looks up the current version of the application.
// GET /version
// Returns the current version of the application.
func GetVersion(w http.ResponseWriter, r *http.Request) {
	var data Version
	configuration := toolbox.Configuration{}

	env := []string{}
	env = append(env, "config/config.", configuration.Environment, ".json")
	err := gonfig.GetConf(strings.Join(env, ""), &configuration)
	if (err != nil) {
		toolbox.RespondWithError(w, http.StatusBadRequest, "Application Version details unknown!")
	}

	data.Version = configuration.Version
	data.ReleaseDate = configuration.ReleaseDate
	toolbox.RespondWithJSON(w, http.StatusOK, data)
}

// main function
// Starting point for application
func main() {

	// API router
    // Consider use of .StrictSlash(true)
	router := mux.NewRouter()

	subRouterAuthors := router.PathPrefix("/authors").Subrouter()
	subRouterAuthor := router.PathPrefix("/author").Subrouter()
	subRouterQuotes := router.PathPrefix("/quotes").Subrouter()
	subRouterQuote := router.PathPrefix("/quote").Subrouter()
	subRouterHealth := router.PathPrefix("/health").Subrouter()
	subRouterReady := router.PathPrefix("/ready").Subrouter()
	subRouterVersion := router.PathPrefix("/version").Subrouter()

	// GET /authors
	subRouterAuthors.HandleFunc("", GetAuthors).Methods("GET")
	subRouterAuthors.HandleFunc("/", GetAuthors).Methods("GET")

	// GET /author
	subRouterAuthor.HandleFunc("/{id}",  GetAuthor).Methods("GET")
	subRouterAuthor.HandleFunc("/{id}/", GetAuthor).Methods("GET")

	// POST /author
	subRouterAuthor.HandleFunc("", CreateAuthor).Methods("POST")
	subRouterAuthor.HandleFunc("/", CreateAuthor).Methods("POST")

	// DELETE /author
	subRouterAuthor.HandleFunc("/{id}",  DeleteAuthor).Methods("DELETE")
	subRouterAuthor.HandleFunc("/{id}/", DeleteAuthor).Methods("DELETE")

	// GET /quotes
	subRouterQuotes.HandleFunc("", GetQuotes).Methods("GET")
	subRouterQuotes.HandleFunc("/", GetQuotes).Methods("GET")

	// GET /quote
	subRouterQuote.HandleFunc("/{id}",  GetQuote).Methods("GET")
	subRouterQuote.HandleFunc("/{id}/", GetQuote).Methods("GET")

	// POST /quote
	subRouterQuote.HandleFunc("", CreateQuote).Methods("POST")
	subRouterQuote.HandleFunc("/", CreateQuote).Methods("POST")

	// DELETE /quote
	subRouterQuote.HandleFunc("/{id}",  DeleteQuote).Methods("DELETE")
	subRouterQuote.HandleFunc("/{id}/", DeleteQuote).Methods("DELETE")

	// GET /health
	subRouterHealth.HandleFunc("", GetHealth).Methods("GET")
	subRouterHealth.HandleFunc("/", GetHealth).Methods("GET")

	// GET /ready
	subRouterReady.HandleFunc("", toolbox.GetReady).Methods("GET")
	subRouterReady.HandleFunc("/", toolbox.GetReady).Methods("GET")

	// GET /version
	subRouterVersion.HandleFunc("", GetVersion).Methods("GET")
	subRouterVersion.HandleFunc("/", GetVersion).Methods("GET")

	if (toolbox.Conf.Port == 0) {
		fmt.Println("Application port setting not found")
		os.Exit(1)
	}
	port := []string{}
	port = append(port, ":", strconv.Itoa(toolbox.Conf.Port))

	fmt.Printf("Starting server on port %s\n", strings.Join(port, ""))
	log.Fatal(http.ListenAndServe(strings.Join(port, ""),
		handlers.LoggingHandler(os.Stdout, handlers.CORS(
			handlers.AllowedMethods([]string{"GET", "POST", "DELETE"}),
			handlers.AllowedOrigins([]string{"*"}))(router))))
}
