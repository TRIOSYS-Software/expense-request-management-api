package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"shwetaik-expense-management-api/sqlacc"
)

// Both expense and advance requests post the same payment-voucher document to
// SQLACC. The two implementations were identical apart from the document-number
// infix ("PV" vs "ADV") and the log tag, so they now share this one path.

const (
	paymentVouchersPath = "/payment-vouchers/direct"

	// docTypeExpense and docTypeAdvance are the infixes in a voucher document
	// number. They are part of the accounting document identity — changing one
	// would orphan every voucher already posted under the old prefix.
	docTypeExpense = "PV"
	docTypeAdvance = "ADV"
)

// voucherRequest is the projection of an expense or advance request that a
// payment voucher is built from.
type voucherRequest struct {
	ID                       uint
	Description              string
	Project                  string
	Amount                   float64
	PaymentMethod            string
	PaymentMethodDescription string
	GLAccountCode            string
}

// voucherDocNo builds the SQLACC document number.
//
// The middle letter records how the money moves: B for a bank payment method, C
// for anything else (cash). The check is a substring match on the payment
// method's description, preserved exactly from the original implementation.
func voucherDocNo(paymentMethodDescription, docType string, id uint) string {
	channel := "C"
	if strings.Contains(strings.ToLower(paymentMethodDescription), "bank") {
		channel = "B"
	}
	return fmt.Sprintf("APP-%s-%s-%d", channel, docType, id)
}

func buildVoucherPayload(req voucherRequest, docType string) map[string]any {
	return map[string]any{
		"docno":         voucherDocNo(req.PaymentMethodDescription, docType, req.ID),
		"docdate":       time.Now().Format("2006-01-02"),
		"paymentmethod": req.PaymentMethod,
		"description":   req.Description,
		"project":       req.Project,
		"docamt":        req.Amount,
		"sdsdocdetail": []map[string]any{
			{
				"code":        req.GLAccountCode,
				"description": req.Description,
				"amount":      req.Amount,
				"project":     req.Project,
			},
		},
	}
}

// sendVoucher posts the voucher and turns a non-2xx response into an error.
// logTag identifies the request family in logs ("ER" or "AR").
func sendVoucher(ctx context.Context, req voucherRequest, docType, logTag string) error {
	body, err := json.Marshal(buildVoucherPayload(req, docType))
	if err != nil {
		return err
	}

	resp, err := sqlacc.Default().Post(ctx, paymentVouchersPath, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		snippet := readBodySnippet(resp.Body, 1024)
		log.Printf("[send-to-sqlacc] %s id=%d payment-vouchers POST -> %d body=%s payload=%s",
			logTag, req.ID, resp.StatusCode, snippet, string(body))
		return fmt.Errorf("payment-vouchers POST failed: status %d body=%s", resp.StatusCode, snippet)
	}
	return nil
}

// readBodySnippet reads at most max bytes of an error response, for logging.
func readBodySnippet(r io.Reader, max int) string {
	if r == nil {
		return ""
	}
	buf, err := io.ReadAll(io.LimitReader(r, int64(max)))
	if err != nil {
		return fmt.Sprintf("<read err: %v>", err)
	}
	s := strings.TrimSpace(string(buf))
	if s == "" {
		return "<empty>"
	}
	return s
}
