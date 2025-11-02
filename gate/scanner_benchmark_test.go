package gate

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// BenchmarkScannerLarge generates many small files and benchmarks the generic scanner.
func BenchmarkScannerLarge(b *testing.B) {
    dir := b.TempDir()
    total := 5000
    for i := 0; i < total; i++ {
        ext := ".go"
        switch i % 5 {
        case 1:
            ext = ".rs"
        case 2:
            ext = ".zig"
        case 3:
            ext = ".tsx"
        case 4:
            ext = ".md"
        }
        content := fmt.Sprintf("// CANARY: REQ=CBIN-%03d; FEATURE=\"Bench\"; ASPECT=API; STATUS=IMPL; OWNER=bench; UPDATED=2025-11-02\n", i)
        if ext == ".md" {
            content = fmt.Sprintf("<!-- CANARY: REQ=CBIN-%03d; FEATURE=\"Bench\"; ASPECT=API; STATUS=IMPL; OWNER=bench; UPDATED=2025-11-02 -->\n", i)
        }
        name := fmt.Sprintf("file_%d%s", i, ext)
        if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
            b.Fatalf("write %s: %v", name, err)
        }
    }
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        sc := NewScanner()
        if _, err := sc.ScanRepository(dir); err != nil {
            b.Fatalf("scan error: %v", err)
        }
    }
}
