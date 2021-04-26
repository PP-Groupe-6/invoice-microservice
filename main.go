package main

import (
	"net/http"
	"os"

	invoiceService "github.com/PP-Groupe-6/invoice-microservice/invoice_microservice"
	"github.com/go-kit/kit/log"
)

func main() {
	info := invoiceService.DbConnexionInfo{
		DbUrl:    "postgre://",
		DbPort:   "5432",
		DbName:   "prix_banque_test",
		Username: "dev",
		Password: "dev",
	}

	service := invoiceService.NewInvoiceService(info)

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	err := http.ListenAndServe(":8000", invoiceService.MakeHTTPHandler(service, logger))
	if err != nil {
		panic(err)
	}
}
