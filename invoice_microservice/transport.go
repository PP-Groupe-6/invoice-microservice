package invoice_microservice

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/transport"
	"github.com/gorilla/mux"
	"github.com/rs/cors"

	httptransport "github.com/go-kit/kit/transport/http"
)

var (
	// ErrBadRouting is returned when an expected path variable is missing.
	// It always indicates programmer error.
	ErrBadRouting = errors.New("inconsistent mapping between route and handler (programmer error)")
)

func MakeHTTPHandler(s InvoiceService, logger log.Logger) http.Handler {
	r := mux.NewRouter()
	e := MakeInvoiceEndpoints(s)
	options := []httptransport.ServerOption{
		httptransport.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
		httptransport.ServerErrorEncoder(encodeError),
	}

	// GET		/invoices/ 		returns the invoices given an account id and the created boolean
	// POST		/invoices/ 		creates an invoice with the given information
	// DELETE 	/invoices/		deletes the invoice corresponding to the given ID
	// POST		/invoices/pay	tries to process the payment of the given invoice

	r.Methods("GET").Path("/invoices/{id}").Handler(httptransport.NewServer(
		e.GetInvoiceListEndpoint,
		decodeInvoiceListRequest,
		encodeResponse,
		options...,
	))

	r.Methods("POST").Path("/invoices/").Handler(httptransport.NewServer(
		e.AddEndpoint,
		decodeAddRequest,
		encodeResponse,
		options...,
	))

	r.Methods("DELETE").Path("/invoices/").Handler(httptransport.NewServer(
		e.DeleteEndpoint,
		decodeDeleteRequest,
		encodeResponse,
		options...,
	))

	r.Methods("POST").Path("/invoices/pay").Handler(httptransport.NewServer(
		e.InvoicePaiementEndpoint,
		decodePayRequest,
		encodeResponse,
		options...,
	))

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"POST", "GET", "OPTIONS"},
		//AllowedHeaders: []string{"Content-Type", "Accept", "Accept-Encoding", "Authorization"},
		AllowedHeaders: []string{"*"},
		// Enable Debugging for testing, consider disabling in production
		Debug: true,
	})

	handler := c.Handler(r)

	return handler
}

func decodeInvoiceListRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	createdBy, _ := strconv.ParseBool(r.URL.Query().Get("CreatedBy"))
	idparam, ok := vars["id"]
	if !ok {
		return nil, ErrBadRouting
	}
	var req = GetInvoiceListRequest{
		idparam,
		createdBy,
	}
	return req, nil
}

func decodeAddRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	var req AddRequest
	if e := json.NewDecoder(r.Body).Decode(&req); e != nil {
		return nil, e
	}
	return req, nil
}

func decodePayRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	var req InvoicePaymentRequest
	if e := json.NewDecoder(r.Body).Decode(&req); e != nil {
		return nil, e
	}
	return req, nil
}

func decodeDeleteRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	var req DeleteRequest
	if e := json.NewDecoder(r.Body).Decode(&req); e != nil {
		return nil, e
	}
	return req, nil
}

type errorer interface {
	error() error
}

func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	if e, ok := response.(errorer); ok && e.error() != nil {
		// Not a Go kit transport error, but a business-logic error.
		// Provide those as HTTP errors.
		encodeError(ctx, e.error(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	if err == nil {
		panic("encodeError with nil error")
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(codeFrom(err))
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}

func codeFrom(err error) int {
	switch err {
	case ErrNotFound:
		return http.StatusNotFound
	case ErrNotAnId, ErrNotFound:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
