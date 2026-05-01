package postgres

import (
	"database/sql/driver"

	"github.com/jackc/pgx/v5/pgtype"
)

/* ------------ Text  ------------ */

type Text struct {
	pgtype.Text
}

func NewText(s string) Text {

	if s == "" {
		return Text{Text: pgtype.Text{Valid: false}}
	}

	return Text{Text: pgtype.Text{String: s, Valid: true}}
}

func (t Text) String() string {
	if !t.Valid {
		return ""
	}

	return t.Text.String
}

func (t Text) Ptr() *string {
	if !t.Valid {
		return nil
	}
	return &t.Text.String
}

func (t Text) IsNull() bool {
	return !t.Valid
}

func (t *Text) Set(s string) {
	t.Text = pgtype.Text{
		String: s,
		Valid:  s != "",
	}
}

func (t *Text) Scan(value any) error {
	return t.Text.Scan(value)
}

func (t Text) Value() (driver.Value, error) {
	return t.Text.Value()
}
