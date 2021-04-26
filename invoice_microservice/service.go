package invoice_microservice

import (
	"context"
	"errors"
	"fmt"

	"github.com/rs/xid"
)

type InvoiceService interface {
	Create(ctx context.Context, invoice Invoice) (Invoice, error)
	Read(ctx context.Context, id string) (Invoice, error)
	Update(ctx context.Context, id string, invoice Invoice) (Invoice, error)
	Delete(ctx context.Context, id string) error
	GetInvoiceList(ctx context.Context, id string) ([]Invoice, error)
	GetIdFromMail(ctx context.Context, mail string) (string, error)
	PayInvoice(ctx context.Context, id string) (bool, error)
	GetAccountInformation(ctx context.Context, id string) (AccountInfo, error)
}

var (
	ErrNotAnId             = errors.New("not an ID")
	ErrNotFound            = errors.New("invoice not found")
	ErrNoTransfer          = errors.New("invoice field is empty")
	ErrNoUpdate            = errors.New("could not update invoice")
	ErrNoDb                = errors.New("could not access database")
	ErrAlreadyExist        = errors.New("invoice id already exists")
	ErrNoInsert            = errors.New("insert did not go through")
	ErrInconsistentIDs     = errors.New("could not access database")
	ErrInsufficientBalance = errors.New("payer's balance is to low to pay invoice")
	ErrAccountNotFound     = errors.New("requested account was not found")
)

type invoiceService struct {
	DbInfos DbConnexionInfo
}

func NewInvoiceService(dbinfos DbConnexionInfo) InvoiceService {
	return &invoiceService{
		DbInfos: dbinfos,
	}
}

func (s *invoiceService) GetInvoiceList(ctx context.Context, id string) ([]Invoice, error) {
	db := GetDbConnexion(s.DbInfos)

	invoices := make([]Invoice, 0)
	rows, err := db.Queryx("SELECT * FROM invoice WHERE account_invoice_payer_id=$1 OR account_invoice_receiver_id=$1", id)

	for rows.Next() {
		var i Invoice
		if err := rows.StructScan(&i); err != nil {
			return nil, err
		}

		invoices = append(invoices, i)
	}

	if err != nil {
		return nil, err
	}

	return invoices, err
}

func (s *invoiceService) Create(ctx context.Context, invoice Invoice) (Invoice, error) {
	if (invoice == Invoice{}) {
		return Invoice{}, ErrNoTransfer
	}

	if testID, _ := s.Read(ctx, invoice.ID); (testID != Invoice{}) {
		return Invoice{}, ErrAlreadyExist
	}

	// Génération d'un UUID
	id := xid.New()

	db := GetDbConnexion(s.DbInfos)
	tx := db.MustBegin()
	res := tx.MustExec("INSERT INTO invoice VALUES('" + id.String() + "','" + fmt.Sprint(invoice.Amount) + "','" + fmt.Sprint(invoice.State) + "','" + invoice.ExpirationDate + "','" + invoice.AccountPayerId + "','" + invoice.AccountReceiverId + "')")
	tx.Commit()
	db.Close()

	if nRows, err := res.RowsAffected(); nRows != 1 || err != nil {
		if err != nil {
			return Invoice{}, err
		}
		return Invoice{}, ErrNoInsert
	}

	invoice.ID = id.String()
	inserted, _ := s.Read(ctx, invoice.ID)

	return inserted, nil
}

func (s *invoiceService) Read(ctx context.Context, id string) (Invoice, error) {
	db := GetDbConnexion(s.DbInfos)

	Res := Invoice{}
	err := db.Get(&Res, "SELECT * FROM invoice WHERE invoice_id=$1", id)

	if err != nil {
		return Invoice{}, err
	}

	return Res, nil
}

func (s *invoiceService) Update(ctx context.Context, id string, invoice Invoice) (Invoice, error) {
	if (invoice == Invoice{}) {
		return Invoice{}, ErrNoTransfer
	}

	if testID, _ := s.Read(ctx, id); (testID == Invoice{}) {
		return Invoice{}, ErrNotFound
	}

	db := GetDbConnexion(s.DbInfos)
	tx := db.MustBegin()
	res := tx.MustExec("UPDATE invoice SET invoice_amount = '"+fmt.Sprint(invoice.Amount)+"', invoice_state ='"+fmt.Sprint(invoice.State)+"', invoice_expiration_date = '"+invoice.ExpirationDate+"', account_invoice_payer_id = '"+invoice.AccountPayerId+"', account_invoice_receiver_id = '"+invoice.AccountReceiverId+"' WHERE invoice_id=$1", id)
	tx.Commit()
	db.Close()

	if nRows, err := res.RowsAffected(); nRows != 1 || err != nil {
		if err != nil {
			return Invoice{}, err
		}
		return Invoice{}, ErrNoInsert
	}

	return s.Read(ctx, invoice.ID)
}

func (s *invoiceService) Delete(ctx context.Context, id string) error {
	if testID, _ := s.Read(ctx, id); (testID == Invoice{}) {
		return ErrNotFound
	}
	db := GetDbConnexion(s.DbInfos)
	tx := db.MustBegin()
	res := tx.MustExec("DELETE FROM invoice WHERE invoice_id=$1", id)

	if nRows, err := res.RowsAffected(); nRows != 1 || err != nil {
		if err != nil {
			return err
		}
	}
	tx.Commit()
	db.Close()

	return nil
}

func (s *invoiceService) GetIdFromMail(ctx context.Context, mail string) (string, error) {
	db := GetDbConnexion(s.DbInfos)

	res := ""
	err := db.Get(&res, "SELECT client_id FROM account WHERE mail_adress=$1", mail)

	if err != nil {
		return "", err
	}

	return res, nil
}

func (s *invoiceService) PayInvoice(ctx context.Context, id string) (bool, error) {
	if id == "" {
		return false, ErrNotAnId
	}

	InvoiceToPay, errorR := s.Read(ctx, id)

	if (InvoiceToPay == Invoice{} && errorR != nil) {
		return false, ErrNotFound
	}

	db := GetDbConnexion(s.DbInfos)

	// Dans un premier temps on récupère le solde du payeur
	payerBalance := float64(0.0)

	fmt.Println(InvoiceToPay.AccountPayerId)
	errPB := db.Get(&payerBalance, "SELECT account_amount FROM account WHERE client_id=$1", InvoiceToPay.AccountPayerId)

	// On récupère ensuite le solde du receveur
	recieverBalance := float64(0.0)
	errRB := db.Get(&recieverBalance, "SELECT account_amount FROM account WHERE client_id=$1", InvoiceToPay.AccountReceiverId)

	if errPB != nil {
		fmt.Println("Payer balance error")
		return false, ErrAccountNotFound
	}
	if errRB != nil {
		fmt.Println("Reciever balance error")
		return false, ErrAccountNotFound
	}

	// On regarde si le payeur a les fonds pour payer la facture
	if payerBalance < InvoiceToPay.Amount {
		return false, ErrInsufficientBalance
	}

	tx := db.MustBegin()
	// On mets à jour le solde du payeur
	resPayer := tx.MustExec("UPDATE account SET account_amount = '"+fmt.Sprint(payerBalance-InvoiceToPay.Amount)+"' WHERE client_id=$1", InvoiceToPay.AccountPayerId)

	if rows, errUpdate := resPayer.RowsAffected(); rows != 1 {
		tx.Rollback()
		return false, errUpdate
	}

	// On mets à jour le solde du receveur
	resReciever := tx.MustExec("UPDATE account SET account_amount = '"+fmt.Sprint(recieverBalance+InvoiceToPay.Amount)+"' WHERE client_id=$1", InvoiceToPay.AccountPayerId)
	if rows, errUpdate := resReciever.RowsAffected(); rows != 1 {
		tx.Rollback()
		return false, errUpdate
	}

	//On change l'état de la facture a payer
	resInvoice := tx.MustExec("UPDATE invoice SET invoice_state = '"+fmt.Sprint(PAID)+"' WHERE invoice_id=$1", InvoiceToPay.ID)
	if rows, errUpdate := resInvoice.RowsAffected(); rows != 1 {
		tx.Rollback()
		return false, errUpdate
	}

	tx.Commit()
	db.Close()

	return true, nil
}

func (s *invoiceService) GetAccountInformation(ctx context.Context, id string) (AccountInfo, error) {
	db := GetDbConnexion(s.DbInfos)

	res := AccountInfo{}
	err := db.Get(&res, "SELECT name, surname, mail_adress, phone_number, account_amount FROM account where client_id=$1", id)

	if err != nil {
		return AccountInfo{}, err
	}
	return res, err
}
