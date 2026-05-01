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

	statementPath := "Statements (6).pdf"

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
	// fmt.Println(" RAW ==================================================================================")
	// fmt.Printf("[piraeus/v2] RAW PDF TO TEXT ------\n%s\n", text)
	// fmt.Println(" END OF RAW ==================================================================================")
	pages := strings.Split(text, "\f")

	// for i, page := range pages {
	// 	fmt.Println("==========================================")
	// 	fmt.Printf("PAGE %d\n", i)
	// 	fmt.Println(page)
	// }

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

			if i == 0 {
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
					fmt.Println("start of transaction")
					// rows = append(rows, cols1)
				}

				transaction = append(transaction, trimmedLine)

			}

			if i == 61 {
				isHeader = false
				isSetHeader = false
			}

		}
	}
	fmt.Println("HEADER LINES ========================================================")
	for _, line := range headerLines {
		fmt.Println(line)
	}
	re := regexp.MustCompile(`\d{4}-\d{6}-\d{3}`)

	accountNumber := re.FindString(headerLines[5])

	fmt.Println(accountNumber)
	fmt.Println("HERE" + headerLines[5] + " account is:" + accountNumber)
	fmt.Println("========================================================")

	// for _, transaction := range transactions {
	// 	fmt.Println("TRANSACTION:")
	// 	for _, col := range transaction {
	// 		fmt.Println(col)
	// 	}
	// }

	previousBalance := headerLines[57] + " " + headerLines[58] + " " + headerLines[59]

	previousBal, _ := toCents(previousBalance)
	//fmt.Printf("Previous balance raw: %q\n", previousBal)

	var parsed []Transaction
	var previousDate time.Time
	var previousBalanceAfter int

	for i, rawTx := range transactions {

		fmt.Printf("Processing transaction %d: %v\n", i, rawTx)
		if strings.Contains(rawTx[1], "Νέ") && strings.Contains(rawTx[2], "οΥπόλοι") {
			continue
		}

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
		} else {
			tx.BalanceBefore = previousBal
			if tx.BalanceAfter < previousBal {
				tx.Direction = "debit"
				tx.Amount = previousBal - tx.BalanceAfter
			} else {
				tx.Direction = "credit"
				tx.Amount = tx.BalanceAfter - previousBal
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

	// for r, row := range rawTx {
	// 	fmt.Printf("row %d: %q\n", r, row)
	// }
	cols := regexp.MustCompile(`\s{2,}`).Split(strings.TrimSpace(firstRow), -1)
	// fmt.Printf("parseTransaction cols: %d\n", len(cols))
	// for i, col := range cols {
	// 	fmt.Printf("col %d: %q\n", i, col)
	// }

	var tx Transaction
	tx.RawData = rawTx
	tx.AccountID = accountNumber
	// Date exists only if first column is dd/mm/yy
	if len(cols) > 0 && isDate(cols[0]) {
		tx.Date = parseDate(cols[0])
		cols = cols[1:]
	}

	if len(cols) > 4 {
		tx.Amount = parseAmount(cols[3])
	} else {
		tx.Amount = parseAmount(cols[2])
	}

	fmt.Printf("After date parsing, cols: %d\n", tx.Amount)
	// tx.Amount = parseAmount(cols[4])

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

	//FIND BALANCE AFTER.......
	//take the last column of the first row.
	//take the second row.
	//take the third row.
	lastColFirstRow := cols[len(cols)-1]

	// combine everything
	result := lastColFirstRow + " " + rawTx[1] + " " + rawTx[2]

	fmt.Println(result)

	val, _ := toCents(result)
	// Last column is balance after

	tx.BalanceAfter = val

	// Transaction reference usually appears in row:
	// ╬ò╬¥╬ö╬ò╬Ö╬₧╬ù: PO24185000335494
	for _, row := range rawTx {
		fmt.Printf("Looking for transaction reference in row: %q\n", row)
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
		tx.MerchantIdentifier = &rawTx[4]
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

func toCents(input string) (int, error) {
	// 1. Remove everything except digits, comma, dot
	re := regexp.MustCompile(`[^\d,\.]`)
	clean := re.ReplaceAllString(input, "")

	// 2. Remove spaces (just in case)
	clean = strings.ReplaceAll(clean, " ", "")

	// 3. Handle formats like "13,836.28"
	// remove thousands separator
	clean = strings.ReplaceAll(clean, ",", "")

	// 4. Convert to float
	f, err := strconv.ParseFloat(clean, 64)
	if err != nil {
		return 0, err
	}

	// 5. Convert to cents
	return int(f * 100), nil
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
