package netrc

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// benchNetrcFile writes a netrc with n machines and returns its path.
func benchNetrcFile(b *testing.B, n int) string {
	b.Helper()
	var sb strings.Builder
	for i := 0; i < n; i++ {
		sb.WriteString("machine host")
		sb.WriteString(strconv.Itoa(i))
		sb.WriteString(" login user password pass\n")
	}
	path := filepath.Join(b.TempDir(), "netrc")
	if err := os.WriteFile(path, []byte(sb.String()), 0600); err != nil {
		b.Fatal(err)
	}
	return path
}

// BenchmarkFindMachine looks up every machine in turn, which is the access
// pattern of a caller resolving credentials for a large host list.
func BenchmarkFindMachine(b *testing.B) {
	for _, n := range []int{100, 1000, 10000} {
		b.Run("machines="+strconv.Itoa(n), func(b *testing.B) {
			nrc, err := ParseFile(benchNetrcFile(b, n))
			if err != nil {
				b.Fatal(err)
			}
			names := make([]string, n)
			for i := range names {
				names[i] = "host" + strconv.Itoa(i)
			}

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				for _, name := range names {
					if m := nrc.FindMachine(name); m == nil {
						b.Fatalf("missing %s", name)
					}
				}
			}
		})
	}
}

// BenchmarkFindMachineMiss measures the worst case, a full scan that finds
// nothing.
func BenchmarkFindMachineMiss(b *testing.B) {
	nrc, err := ParseFile(benchNetrcFile(b, 10000))
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if m := nrc.FindMachine("absent"); m != nil {
			b.Fatal("unexpected match")
		}
	}
}
