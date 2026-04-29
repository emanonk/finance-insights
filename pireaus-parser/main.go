package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Transaction struct {
	AccountID            string    `json:"accountId"`
	Date                 time.Time `json:"date"`
	BankReference        *string   `json:"bankReference"`
	TransactionReference *string   `json:"transactionReference"`
	MerchantIdentifier   *string   `json:"merchantIdentifier"`
	BalanceBefore        int       `json:"balanceBefore"`
	BalanceAfter         int       `json:"balanceAfter"`
	Amount               int       `json:"amount"`
	Direction            string    `json:"direction"` // debit / credit
	RawData              []string  `json:"rawData"`
}

func main() {
	// if len(os.Args) < 2 {
	// 	panic("usage: go run main.go statement.pdf")
	// }

	statementPath := "Statements (4).pdf"

	// rawBytes, err := os.ReadFile(statementPath)
	// if err != nil {
	// 	panic(err)
	// }

	// decoder := charmap.CodePage850.NewDecoder()

	// fixedBytes, err := decoder.Bytes(rawBytes)
	// if err != nil {
	// 	panic(err)
	// }

	// text := string(fixedBytes)

	text, err := extractText(statementPath)
	if err != nil {
		panic(err)
	}

	pages := strings.Split(text, "\f")

	for i, page := range pages {
		fmt.Println("==========================================")
		fmt.Printf("PAGE %d\n", i)
		fmt.Println(page)
	}

	// var currentDate string = ""
	// var transactionIdx int = 0
	var headerLines []string
	isSetHeader := true
	var transactions [][]string

	fmt.Println(" LINE BY LINE ==================================================================================")
	for pageIndex, page := range pages {
		// fmt.Println("====================================================================================================")
		// fmt.Printf("PAGE %d\n", i)
		// fmt.Println(page)

		lines := strings.Split(strings.ReplaceAll(page, "\r\n", "\n"), "\n")
		isHeader := false

		var transaction []string
		for i, line := range lines {
			fmt.Printf("==================================================page %d/%d\n", pageIndex, i)

			fmt.Printf("---%q-\n", line)
			trimmedLine := strings.TrimSpace(string(line))
			fmt.Printf("---trim--%q-\n", trimmedLine)

			if trimmedLine == "1" {
				isHeader = true
			}

			if isHeader && isSetHeader {
				headerLines = append(headerLines, line)
				fmt.Printf("$header-row :%s\n", line)
			}

			if !isHeader {
				fmt.Printf("$trans-row :%s\n", line)

				re := regexp.MustCompile(`\s{2,}`)

				cols1 := re.Split(strings.TrimSpace(line), -1)

				fmt.Println("Line 1 columns:", len(cols1))
				fmt.Println(cols1)

				if len(cols1) >= 4 || trimmedLine == "" {
					// save previous transaction first (if not empty)
					if len(transaction) > 2 {
						details := "PAGE:" + fmt.Sprint(pageIndex) + " LINE:" + fmt.Sprint(i)
						transaction = append(transaction, details)
						transactions = append(transactions, transaction)
					}

					transaction = nil // start new transaction
					// rows = append(rows, cols1)
				}

				transaction = append(transaction, trimmedLine)

			}

			if strings.Contains(trimmedLine, "Ðñïçãïýìåíï Õðüëïéðï") {
				isHeader = false
				isSetHeader = false
			}

		}
	}
	fmt.Println("========================================================")
	for _, line := range headerLines {
		fmt.Println(line)
	}
	re := regexp.MustCompile(`\d{4}-\d{6}-\d{3}`)

	accountNumber := re.FindString(headerLines[12])

	fmt.Println(accountNumber)
	fmt.Println("HERE" + headerLines[12] + " account is:" + accountNumber)
	fmt.Println("========================================================")

	for _, transaction := range transactions {
		fmt.Println("TRANSACTION:")
		for _, col := range transaction {
			fmt.Println(col)
		}
	}

	var parsed []Transaction
	var previousDate time.Time
	var previousBalanceAfter int

	for i, rawTx := range transactions {
		tx := parseTransaction(rawTx, accountNumber)

		// If first row has no date, reuse previous transaction date
		if tx.Date.IsZero() {
			tx.Date = previousDate
		} else {
			previousDate = tx.Date
		}

		// Compare with previous transaction balance
		if i > 0 {
			tx.BalanceBefore = previousBalanceAfter

			if tx.BalanceAfter < previousBalanceAfter {
				tx.Direction = "debit"
				tx.Amount = previousBalanceAfter - tx.BalanceAfter
			} else {
				tx.Direction = "credit"
				tx.Amount = tx.BalanceAfter - previousBalanceAfter
			}
		}

		previousBalanceAfter = tx.BalanceAfter

		parsed = append(parsed, tx)
	}

	jsonData, err := json.MarshalIndent(parsed, "", "  ")
	if err != nil {
		fmt.Println("json error:", err)
		return
	}

	fmt.Println(string(jsonData))

	// fmt.Println(transactions)

}

func parseTransaction(rawTx []string, accountNumber string) Transaction {
	firstRow := rawTx[0]

	cols := regexp.MustCompile(`\s{2,}`).Split(strings.TrimSpace(firstRow), -1)

	var tx Transaction
	tx.RawData = rawTx
	tx.AccountID = accountNumber
	// Date exists only if first column is dd/mm/yy
	if len(cols) > 0 && isDate(cols[0]) {
		tx.Date = parseDate(cols[0])
		cols = cols[1:]
	}

	// Example:
	// 2960 EL01P 0442174
	if len(cols) > 0 {
		// parts := strings.Fields(cols[0])
		tx.BankReference = &cols[0]
		// if len(parts) >= 3 {
		// 	ref := parts[2]
		// 	tx.BankReference = &ref
		// }
	}

	// Last column is balance after
	if len(cols) > 0 {
		tx.BalanceAfter = parseAmount(cols[len(cols)-1])
	}

	// Transaction reference usually appears in row:
	// ╬ò╬¥╬ö╬ò╬Ö╬₧╬ù: PO24185000335494
	for _, row := range rawTx {
		if strings.Contains(row, ":") {
			parts := strings.SplitN(row, ":", 2)
			ref := strings.TrimSpace(parts[1])
			if ref != "" {
				tx.TransactionReference = &ref
				break
			}
		}
	}
	if len(rawTx) > 2 {
		tx.MerchantIdentifier = &rawTx[2]
	}

	return tx
}

func extractText(pdfPath string) (string, error) {
	cmd := exec.Command("pdftotext", "-layout", "-enc", "UTF-8", pdfPath, "-")
	out, err := cmd.Output()
	// fmt.Printf("[piraeus/v1] RAW PDF TO TEXT ------\n%s\n", out)
	return string(out), err
}

func isDate(value string) bool {
	_, err := time.Parse("02/01/06", value)
	return err == nil
}

func parseDate(value string) time.Time {
	t, _ := time.Parse("02/01/06", value)
	return t
}

func parseAmount(value string) int {
	value = strings.TrimSpace(value)

	// remove garbage suffix like ├É├ë
	value = regexp.MustCompile(`[^0-9,.-]`).ReplaceAllString(value, "")

	// 1,987.26 -> 1987.26
	value = strings.ReplaceAll(value, ".", "") // remove thousand separators if any
	value = strings.ReplaceAll(value, ",", "")

	amount, _ := strconv.Atoi(value)
	return amount
}
