package reserve

import (
	"encoding/json"
	"gopkg.in/go-playground/validator.v9"
)

type Reserve struct {
	ID                int64   `json:"id"`
	Version           *string `json:"version"`
	TTL               *int64  `json:"-"`
	ExternalReference string  `json:"-"`
	IdempotencyKey    string  `json:"-"`
	Reason            Reason  `json:"-"`
	Mode              Mode    `json:"-"`
	Amount            int64   `json:"amount"`
	ClientID          string  `json:"-"`
	UserID            uint64  `json:"-"`
	Status            string  `json:"status"`
	DateCreated       string  `json:"-"`
	LastModified      string  `json:"-"`
}

type ByAmount []Reserve

func (a ByAmount) Len() int           { return len(a) }
func (a ByAmount) Less(i, j int) bool { return a[i].Amount > a[j].Amount }
func (a ByAmount) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func ByAmountComparator(a, b interface{}) int {
	ra := a.(Reserve)
	rb := b.(Reserve)

	switch {
	case ra.Amount < rb.Amount:
		return 1
	case ra.Amount > rb.Amount:
		return -1
	default:
		return 0
	}
}

type CreateHeader struct {
	IdempotencyKey string `header:"X-Idempotency-Key" binding:"required"`
	ClientID       string `header:"X-Client-Id"`
}

type CreateQuery struct {
	ClientID string `query:"client.id"`
}

type CreateURI struct {
	UserID uint64 `uri:"user_id" binding:"required"`
}

type Body struct {
	Amount            int64  `json:"amount" binging:"required,gt=0"`
	Mode              Mode   `json:"mode" binging:"required"`
	Reason            Reason `json:"reason" binging:"required"`
	ExternalReference string `json:"external_reference" binging:"required"`
}

type ReserveRequest struct {
	Body           Body
	ClientID       string
	UserID         uint64
	IdempotencyKey string
}

func BodyStructValidation(structLevel validator.StructLevel) {
	reserveBody := structLevel.Current().Interface().(Body)

	if !reserveBody.Reason.Valid() {
		structLevel.ReportError(
			reserveBody.Reason, "reason",
			"Reason", "invalid_reason", "",
		)
	}

	if !reserveBody.Mode.Valid() {
		structLevel.ReportError(
			reserveBody.Mode, "mode",
			"Mode", "invalid_mode", "",
		)
	}
}

func (rb *Body) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Amount            float64 `json:"amount"`
		Mode              string  `json:"mode"`
		Reason            string  `json:"reason"`
		ExternalReference string  `json:"external_reference"`
	}{
		Amount:            float64(rb.Amount) / 100,
		Mode:              string(rb.Mode),
		Reason:            string(rb.Reason),
		ExternalReference: rb.ExternalReference,
	})
}

func (rb *Body) UnmarshalJSON(data []byte) error {
	type Alias Body
	aux := &struct {
		Amount float64 `json:"amount"`
		*Alias
	}{
		Alias: (*Alias)(rb),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	rb.Amount = int64(aux.Amount * 100)
	return nil
}

type Mode string

var Modes = struct {
	Total   Mode
	Partial Mode
}{
	"total",
	"partial",
}

var PossibleModes = []Mode{Modes.Total, Modes.Partial}

func (r *Mode) Valid() bool {
	for _, pr := range PossibleModes {
		if *r == pr {
			return true
		}
	}
	return false
}

type Reason string

var Reasons = struct {
	ReserveForPayment Reason
}{
	"reserve_for_payment",
}

var PossibleReasons = []Reason{Reasons.ReserveForPayment}

func (r *Reason) Valid() bool {
	for _, pr := range PossibleReasons {
		if *r == pr {
			return true
		}
	}
	return false
}
