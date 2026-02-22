package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRespondJSON_EncodingError(t *testing.T) {
	rec := httptest.NewRecorder()
	respondJSON(rec, http.StatusOK, make(chan int))

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}
