package gate

import "testing"

func TestRegisterCommentStyleOverride(t *testing.T) {
	original := DetectStyleFromExtension("file.xyz")
	if original.LinePrefix != "//" {
		t.Fatalf("unexpected default fallback prefix: %s", original.LinePrefix)
	}
	RegisterCommentStyle(".xyz", CommentStyle{LinePrefix: "#"})
	style := DetectStyleFromExtension("sample.xyz")
	if style.LinePrefix != "#" {
		t.Fatalf("expected overridden style '#', got %q", style.LinePrefix)
	}
}
