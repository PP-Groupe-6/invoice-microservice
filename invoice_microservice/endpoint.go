package invoice_microservice

import (
	"context"

	"github.com/go-kit/kit/endpoint"
)

const (
	PENDING = 0
	PAID    = 1
	EXPIRED = 2
)

type InvoiceEndpoints struct {
	GetInvoiceListEndpoint  endpoint.Endpoint
	AddEndpoint             endpoint.Endpoint
	DeleteEndpoint          endpoint.Endpoint
	InvoicePaiementEndpoint endpoint.Endpoint
}

func MakeInvoiceEndpoints(s InvoiceService) InvoiceEndpoints {
	return InvoiceEndpoints{
		GetInvoiceListEndpoint:  MakeGetInvoiceListEndpoint(s),
		AddEndpoint:             MakeAddEndpoint(s),
		DeleteEndpoint:          MakeDeleteEndpoint(s),
		InvoicePaiementEndpoint: MakeInvoicePaymentEndpoint(s),
	}
}

// Si created by est à true on retourne les invoices créées par le client si il est à false on retourne celles reçues par le client
type GetInvoiceListRequest struct {
	ClientID  string
	createdBy bool
}

type GetInvoiceListResponse struct {
	Invoices []InvoiceResponseFormat `json:"invoices"`
}

type InvoiceResponseFormat struct {
	Id           string  `json:"id"`
	Amount       float32 `json:"amount"`
	State        string  `json:"state"`
	ExpDate      string  `json:"expDate"`
	WithClientId string  `json:"withClientId"`
}

func MakeGetInvoiceListEndpoint(s InvoiceService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(GetInvoiceListRequest)
		var invoicesRet []InvoiceResponseFormat
		invoices, err := s.GetInvoiceList(ctx, req.ClientID)

		for _, invoice := range invoices {
			// Si on veut les invoice créées et que l'utilisateur est le récepteur de l'invoice
			if req.createdBy && invoice.AccountReceiverId == req.ClientID {
				invoicesRet = append(invoicesRet, InvoiceResponseFormat{
					invoice.ID,
					float32(invoice.Amount),
					StateToString(invoice.State),
					invoice.ExpirationDate,
					invoice.AccountPayerId,
				})
			}
			// Si on veut les invoice reçues et que l'utilisateur et le payeur de l'invoice
			if !req.createdBy && invoice.AccountPayerId == req.ClientID {
				invoicesRet = append(invoicesRet, InvoiceResponseFormat{
					invoice.ID,
					float32(invoice.Amount),
					StateToString(invoice.State),
					invoice.ExpirationDate,
					invoice.AccountReceiverId,
				})
			}
		}

		return GetInvoiceListResponse{invoicesRet}, err
	}
}

type AddRequest struct {
	Uid         string  // Id du client créant la facture
	EmailClient string  // email du client payeur
	Amount      float32 // montant de la facture
	ExpDate     string  // date d'expiration de la facture
}

type AddResponse struct {
	Created bool `json:"created"`
}

func MakeAddEndpoint(s InvoiceService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(AddRequest)

		id, err := s.GetIdFromMail(ctx, req.EmailClient)

		if err != nil {
			return nil, err
		}

		i := Invoice{
			"",
			float64(req.Amount),
			PENDING,
			req.ExpDate,
			id,
			req.Uid,
		}

		_, err = s.Create(ctx, i)

		if err != nil {
			return AddResponse{true}, nil
		} else {
			return nil, err
		}
	}
}

type InvoicePaymentRequest struct {
	Iid string
}

type InvoicePaymentResponse struct {
	Paid bool `json:"paid"`
}

func MakeInvoicePaymentEndpoint(s InvoiceService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(InvoicePaymentRequest)

		paid, err := s.PayInvoice(ctx, req.Iid)

		return InvoicePaymentResponse{paid}, err
	}
}

type DeleteRequest struct {
	Iid string
}

type DeleteResponse struct {
}

func MakeDeleteEndpoint(s InvoiceService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(DeleteRequest)

		err := s.Delete(ctx, req.Iid)

		return DeleteResponse{}, err
	}
}

func StateToString(stateID int) string {
	switch stateID {
	case PENDING:
		return "Pending"
	case PAID:
		return "Paid"
	case EXPIRED:
		return "Expired"
	}
	return ""
}
