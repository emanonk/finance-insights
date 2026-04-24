package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"

	"statement-parser/parser"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: statement-parser <pdf-file> [output.csv]")
		os.Exit(1)
	}

	pdfFile := os.Args[1]
	outputFile := "transactions.csv"
	if len(os.Args) >= 3 {
		outputFile = os.Args[2]
	}

	p := parser.NewParser()
	txs, err := p.Parse(pdfFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fileName := filepath.Base(pdfFile)
	if err := exportCSV(txs, outputFile, fileName); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing CSV: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Parsed %d transactions → %s\n", len(txs), outputFile)
}

func exportCSV(txs []parser.Transaction, path, fileName string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// UTF-8 BOM for Excel compatibility with Greek characters
	f.Write([]byte{0xEF, 0xBB, 0xBF})

	w := csv.NewWriter(f)
	header := []string{
		"Id",
		"Account_id",
		"Date",
		"Bank_reference_number",
		"Justification",
		"Indicator",
		"merchant_Identifier",
		"Amount1",
		"MCC_Code",
		"Card_masked",
		"Reference",
		"description",
		"Payment_Method",
		"direction",
		"Amount",
		"Balance_after_transaction",
		"Statement_file_name",
	}
	if err := w.Write(header); err != nil {
		return err
	}

	for i, tx := range txs {
		row := []string{
			fmt.Sprintf("%d", i+1),
			tx.AccountID,
			tx.Date,
			tx.BankReferenceNumber,
			tx.Justification,
			tx.Indicator,
			tx.MerchantIdentifier,
			tx.Amount1,
			tx.MCCCode,
			tx.CardMasked,
			tx.Reference,
			tx.Description,
			tx.PaymentMethod,
			tx.Direction,
			tx.Amount,
			tx.BalanceAfterTransaction,
			fileName,
		}
		if err := w.Write(row); err != nil {
			return err
		}
	}

	w.Flush()
	return w.Error()
}
