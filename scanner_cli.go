package canary

// CANARY: REQ=CBIN-101; FEATURE="ScannerCore"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CBIN_101_Engine_ScanBasic; BENCH=BenchmarkCANARY_CBIN_101_Engine_Scan; OWNER=canary; UPDATED=2025-10-15
import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Run executes a scan and writes JSON/CSV outputs. Caller supplies already-built report.
// This separates CLI main from library logic.
func Run(rep Report, out, csv string) error {
	rep.GeneratedAt = time.Now().UTC()
	return writeOutputs(rep, out, csv)
}

func writeOutputs(rep Report, out, csv string) error {
	jf, err := os.Create(out)
	if err != nil {
		return err
	}
	defer func() { _ = jf.Close() }()
	enc := json.NewEncoder(jf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(rep); err != nil {
		return err
	}
	if csv != "" {
		if err := WriteCSV(rep, csv); err != nil {
			return err
		}
	}
	fmt.Printf("CANARY_OK wrote %s\n", out)
	return nil
}
