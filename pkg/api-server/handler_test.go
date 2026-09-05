package api_server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/emicklei/go-restful/v3"
	. "github.com/onsi/gomega"

	rest_error_types "github.com/kumahq/kuma/v3/pkg/core/rest/errors/types"
)

func TestHandleSuccessContracts(t *testing.T) {
	tests := []struct {
		name        string
		body        any
		status      int
		contentType string
		response    string
	}{
		{
			name:        "JSON response",
			body:        map[string]string{"message": "ok"},
			status:      http.StatusOK,
			contentType: "application/json",
			response:    "{\n \"message\": \"ok\"\n}\n",
		},
		{
			name:        "created response",
			body:        created(map[string]string{"message": "created"}),
			status:      http.StatusCreated,
			contentType: "application/json",
			response:    "{\n \"message\": \"created\"\n}\n",
		},
		{
			name:        "raw response",
			body:        rawResponse{contentType: "text/plain", body: []byte("raw body")},
			status:      http.StatusOK,
			contentType: "text/plain",
			response:    "raw body",
		},
		{
			name:        "nil response",
			body:        nil,
			status:      http.StatusOK,
			contentType: "",
			response:    "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)
			response := recordHandlerResponse(func(*restful.Request) (any, error) {
				return test.body, nil
			})

			g.Expect(response.Code).To(Equal(test.status))
			g.Expect(response.Header().Get(restful.HEADER_ContentType)).To(Equal(test.contentType))
			g.Expect(response.Body.String()).To(Equal(test.response))
		})
	}
}

func TestHandleErrorContracts(t *testing.T) {
	tests := []struct {
		name     string
		response any
		err      error
		status   int
		title    string
		detail   string
	}{
		{
			name:     "ordinary error",
			response: map[string]string{"must": "not be written"},
			err:      errors.New("database unavailable"),
			status:   http.StatusInternalServerError,
			detail:   "Internal Server Error",
		},
		{
			name:     "titled error",
			response: map[string]string{"must": "not be written"},
			err:      withTitle(errors.New("database unavailable"), "Could not list resources"),
			status:   http.StatusInternalServerError,
			title:    "Could not list resources",
			detail:   "Internal Server Error",
		},
		{
			name: "typed error title takes precedence",
			err: withTitle(
				fmt.Errorf("wrapped: %w", &rest_error_types.Error{
					Status: http.StatusTeapot,
					Title:  "Typed title",
					Detail: "typed detail",
				}),
				"Ignored title",
			),
			status: http.StatusTeapot,
			title:  "Typed title",
			detail: "typed detail",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)
			response := recordHandlerResponse(func(*restful.Request) (any, error) {
				return test.response, test.err
			})

			g.Expect(response.Code).To(Equal(test.status))
			g.Expect(response.Header().Get(restful.HEADER_ContentType)).To(Equal("application/json"))
			g.Expect(response.Body.String()).ToNot(ContainSubstring("must"))

			body := rest_error_types.Error{}
			g.Expect(json.Unmarshal(response.Body.Bytes(), &body)).To(Succeed())
			g.Expect(body.Status).To(Equal(test.status))
			g.Expect(body.Title).To(Equal(test.title))
			g.Expect(body.Detail).To(Equal(test.detail))
		})
	}
}

func TestWithTitle(t *testing.T) {
	g := NewWithT(t)
	g.Expect(withTitle(nil, "ignored")).To(Succeed())

	sentinel := errors.New("sentinel")
	err := withTitle(sentinel, "Custom title")
	g.Expect(err).To(MatchError("sentinel"))
	g.Expect(errors.Is(err, sentinel)).To(BeTrue())

	var titled *titledError
	g.Expect(errors.As(err, &titled)).To(BeTrue())
	g.Expect(titled.title).To(Equal("Custom title"))
}

func recordHandlerResponse(fn handlerFunc) *httptest.ResponseRecorder {
	request := restful.NewRequest(httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", http.NoBody))
	recorder := httptest.NewRecorder()
	handle(fn)(request, restful.NewResponse(recorder))
	return recorder
}
