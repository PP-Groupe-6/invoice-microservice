package main

import (
	"net/http"
	"os"

	invoiceService "github.com/PP-Groupe-6/invoice-microservice"
	"github.com/go-kit/kit/log"
)

func main() {
	info := invoiceService.dbConnexionInfo{
		"postgre://",
		"5432",
		"prix_banque_test",
		"admin",
		"secret",
	}

	service := invoiceService.NewInvoiceService(info)

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	err := http.ListenAndServe(":8000", MakeHTTPHandler(service, logger))
	if err != nil {
		panic(err)
	}
}
