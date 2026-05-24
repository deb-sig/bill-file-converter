package adapters

import (
	"fmt"
	"sort"
)

type Adapter struct {
	// Key is required. It is the stable CLI --type value and must be unique.
	Key string
	// Name is required. It is the human-readable bill type shown in logs and
	// list-types.
	Name string
	// Headers is required. It is the canonical transaction table header exported
	// to result files. MinerU/VLM output must match either Headers or one
	// HeaderAliases entry before any rows are accepted.
	Headers []string
	// HeaderAliases is optional. Use it only for known MinerU/VLM variants of the
	// same source table header, such as stamp text merged into a header cell or
	// bilingual header rows. When an alias is matched, the exported header is
	// still Headers.
	HeaderAliases [][]string
	// RowGuards is strongly recommended for transaction tables. It identifies
	// real data rows after the header using structural checks, usually on a date
	// or sequence-number column. Do not use it for business-rule cleanup.
	RowGuards []RowGuard
	// BlankRowspanCarryoverColumns is a special-case cleanup knob and should be
	// left unset unless the original bank table visually spans a value across
	// later rows but those later rows semantically have empty cells. Columns are
	// zero-based in Headers; only values carried over from an HTML rowspan are
	// blanked, while the original source cell is preserved.
	BlankRowspanCarryoverColumns []int
	// RemoveImages is an optional preprocessing workaround. Enable it only for
	// profiles where image overlays such as watermarks are known to degrade
	// extraction; leave it off otherwise because PDF rewriting can alter document
	// structure.
	RemoveImages bool
}

type RowGuard struct {
	// Column is zero-based in the canonical table.
	Column int
	// Format is one of the RowGuardFormat* constants below.
	Format RowGuardFormat
}

type RowGuardFormat string

const (
	RowGuardFormatYYYYMMDD               RowGuardFormat = "YYYYMMDD"
	RowGuardFormatYYYYMMDDHHMMSS         RowGuardFormat = "YYYYMMDDHH:mm:ss"
	RowGuardFormatYYYYDashMMDashDDHHMMSS RowGuardFormat = "YYYY-MM-DD HH:mm:ss"
	RowGuardFormatYYYYDashMMDashDD       RowGuardFormat = "YYYY-MM-DD"
	RowGuardFormatMMSlashDD              RowGuardFormat = "MM/DD"
	RowGuardFormatPositiveInteger        RowGuardFormat = "positive_integer"
)

type registry struct {
	adapters map[string]Adapter
}

// NewRegistry creates an adapter registry for Convert callers that provide
// private profiles outside the built-in CLI profile set.
func NewRegistry(adapters ...Adapter) *registry {
	values := map[string]Adapter{}
	for _, adapter := range adapters {
		values[adapter.Key] = adapter
	}
	return &registry{adapters: values}
}

func BuiltinRegistry() *registry {
	return NewRegistry(
		abcDebitAdapter(),
		bocDebitAdapter(),
		bocCreditRegularAdapter(),
		bocCreditReissueAdapter(),
		bocomCreditRegularAdapter(),
		bocomCreditReissueAdapter(),
		bocomDebitAdapter(),
		cmbCreditAdapter(),
		cmbDebitAdapter(),
		zgcDebitAdapter(),
		zbankDebitAdapter(),
	)
}

func (r *registry) List() []Adapter {
	list := make([]Adapter, 0, len(r.adapters))
	for _, adapter := range r.adapters {
		list = append(list, adapter)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Key < list[j].Key
	})
	return list
}

func (r *registry) MustGet(key string) (Adapter, error) {
	if key == "" {
		return Adapter{}, fmt.Errorf("missing bill type: pass --type <adapter_key>; run list-types to see supported types")
	}
	adapter, ok := r.adapters[key]
	if !ok {
		return Adapter{}, fmt.Errorf("unsupported bill type %q: no cleaning profile is registered", key)
	}
	return adapter, nil
}
