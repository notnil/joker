package tournament

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

type icmCase struct {
	Name      string `json:"name"`
	Stacks    []int  `json:"stacks"`
	Payouts   []int  `json:"payouts"`
	Expect    []int  `json:"expect"`
	ExpectSum int    `json:"expectSum"`
	MinEach   int    `json:"minEach"`
	MaxEach   int    `json:"maxEach"`
}

func TestICMCorpus(t *testing.T) {
	var cases []icmCase
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("caller")
	}
	path := filepath.Join(filepath.Dir(file), "..", "..", "testdata", "tournament", "icm.json")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, &cases); err != nil {
		t.Fatal(err)
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			got := ICM(tc.Stacks, tc.Payouts)
			if len(tc.Expect) > 0 {
				for i, w := range tc.Expect {
					if got[i] != w {
						t.Fatalf("got %v, want %v", got, tc.Expect)
					}
				}
			}
			sum := sumInt(got)
			wantSum := tc.ExpectSum
			if wantSum == 0 {
				wantSum = sumInt(tc.Payouts)
			}
			if sum != wantSum {
				t.Fatalf("sum %d, want %d (%v)", sum, wantSum, got)
			}
			if tc.MinEach > 0 {
				for _, v := range got {
					if v < tc.MinEach || (tc.MaxEach > 0 && v > tc.MaxEach) {
						t.Fatalf("value %d outside [%d,%d] in %v", v, tc.MinEach, tc.MaxEach, got)
					}
				}
			}
		})
	}
}
