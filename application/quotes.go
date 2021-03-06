// The "quotes" response functionality for requests to the /quotes endpoint.
// A part of the  quotes methods for the rest-api-go application.
// Governed by the license that can be found in the LICENSE file
package application

import (
	"fmt"
	"net/http"

	"github.com/deezone/rest-api-go/toolbox"
)

// init - one time initialization logic
func init() {
	fmt.Println("- application/quotes rest-api-go package initialized")
}

// GetQuotes looks up all of the quotes.
// GET /quotes
// Returns all of the quotes in JSON format.
func (a *App) GetQuotes(w http.ResponseWriter, r *http.Request) {
	count := 0
	quotes := []Quote{}
	a.DB.Find(&quotes).Count(&count)
	if count == 0 {
		toolbox.RespondWithError(w, http.StatusNotFound, "Quote records not found.")
		return
	}

	// @todo: ISSUE-27 : combine queries with LEFT JOIN
	// SELECT quotes.id AS quote_id, quote, authors.id AS author_id, first, last, born, died, description, bio_link
	// FROM quotes
	//   LEFT JOIN authors ON quotes.author_id = authors.id AND authors.deleted_at IS NULL;

	// Lookup quote author
	// @todo: ISSUE-16 : create parameter to trigger author lookup rather than being the default response
	authormin := AuthorMin{}
	for index, quote := range quotes {
		a.DB.Raw("SELECT * FROM authors WHERE id = ? AND deleted_at IS NULL", quote.AuthorID).Scan(&authormin)
		quotes[index].Author = authormin
	}

	toolbox.RespondWithJSON(w, http.StatusOK, quotes)
}
