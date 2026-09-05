package lexer

import (
	"reflect"
	"testing"
)

func TestPreprocessSingleLineCommentMarkers(t *testing.T) {
	t.Parallel()

	source := "keep me ignored\n// x = 1\n!! input x\nignored too\n"

	got := mustPreprocess(t, source)
	want := []string{"x = 1", "input x"}

	assertLinesEqual(t, got, want)
}

func TestPreprocessLeadingWhitespaceBeforeCommentMarkers(t *testing.T) {
	t.Parallel()

	source := "  // x = 1\n\t!! input x\n\t  //\n"

	got := mustPreprocess(t, source)
	want := []string{"x = 1", "input x", ""}

	assertLinesEqual(t, got, want)
}

func TestPreprocessIgnoresInlineCommentMarkersNotAtLineStart(t *testing.T) {
	t.Parallel()

	source := "x = 1 // not executable\nfoo !! also not executable\n"

	got := mustPreprocess(t, source)
	want := []string{}

	assertLinesEqual(t, got, want)
}

func TestPreprocessBlockCommentSlashStar(t *testing.T) {
	t.Parallel()

	source := "ignored\n/*\nx = 10\ny = 20\n*/\nignored\n"

	got := mustPreprocess(t, source)
	want := []string{"x = 10", "y = 20"}

	assertLinesEqual(t, got, want)
}

func TestPreprocessBlockCommentBangStar(t *testing.T) {
	t.Parallel()

	source := "ignored\n!*\na = 1\nb = 2\n*!\nignored\n"

	got := mustPreprocess(t, source)
	want := []string{"a = 1", "b = 2"}

	assertLinesEqual(t, got, want)
}

func TestPreprocessMixedCommentStyles(t *testing.T) {
	t.Parallel()

	source := "// top\n!! second\n/*\nthird\n*/\n!*\nfourth\n*!\n"

	got := mustPreprocess(t, source)
	want := []string{"top", "second", "third", "fourth"}

	assertLinesEqual(t, got, want)
}

func TestPreprocessPreservesBlankExecutableLines(t *testing.T) {
	t.Parallel()

	source := "//\n// x = 1\n!!\n"

	got := mustPreprocess(t, source)
	want := []string{"", "x = 1", ""}

	assertLinesEqual(t, got, want)
}

func TestPreprocessBlockCommentMultilineKeepsEachLine(t *testing.T) {
	t.Parallel()

	source := "/*\nline one\n\nline three\n*/\n"

	got := mustPreprocess(t, source)
	want := []string{"line one", "", "line three"}

	assertLinesEqual(t, got, want)
}

func TestPreprocessBlockCommentsDoNotNest(t *testing.T) {
	t.Parallel()

	source := "/*\nouter start\n/* this does not open nested\nouter end\n*/\noutside\n"

	got := mustPreprocess(t, source)
	// first */ closes the block; nested opener is inert and dropped;
	// outside is not executable
	want := []string{"outer start", "outer end"}

	assertLinesEqual(t, got, want)
}

func TestPreprocessUnclosedBlockCommentConsumesUntilEOF(t *testing.T) {
	t.Parallel()

	source := "ignore\n/*\na\nb\n"

	got, err := Preprocess(source)
	if err == nil {
		t.Fatal("expected unterminated block comment error")
	}
	want := []string{"a", "b"}

	assertLinesEqual(t, got, want)
}

func TestPreprocessBlockCommentOpenAndCloseSameLine(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		source string
		want   []string
	}{
		{
			name:   "slash-star",
			source: "/* x = 1 */\n",
			want:   []string{"x = 1"},
		},
		{
			name:   "bang-star",
			source: "!* y = 2 *!\n",
			want:   []string{"y = 2"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := mustPreprocess(t, tc.source)
			assertLinesEqual(t, got, tc.want)
		})
	}
}

func TestPreprocessBlockCommentWithLeadingWhitespaceMarker(t *testing.T) {
	t.Parallel()

	source := "   /* x = 1 */\n\t!* y = 2 *!\n"

	got := mustPreprocess(t, source)
	want := []string{"x = 1", "y = 2"}

	assertLinesEqual(t, got, want)
}

func TestPreprocessBlockCommentInnerLineTrimmed(t *testing.T) {
	t.Parallel()

	source := "/*\n   x = 1   \n\t  y = 2\t\n*/\n"

	got := mustPreprocess(t, source)
	want := []string{"x = 1", "y = 2"}

	assertLinesEqual(t, got, want)
}

func TestPreprocessBlockCommentInnerLineWhitespaceContract(t *testing.T) {
	t.Parallel()

	source := "/*\n  line one  \n*/\n"

	got := mustPreprocess(t, source)
	want := []string{"line one"}

	assertLinesEqual(t, got, want)
}

func TestPreprocessCRLF(t *testing.T) {
	t.Parallel()

	source := "// a\r\n!! b\r\n/*\r\nc\r\n*/\r\n"

	got := mustPreprocess(t, source)
	want := []string{"a", "b", "c"}

	assertLinesEqual(t, got, want)
}

func TestPreprocessWindowsPathLikeLineNotComment(t *testing.T) {
	t.Parallel()

	source := `C:\\temp\\file // not marker at line start` + "\n"

	got := mustPreprocess(t, source)
	want := []string{}

	assertLinesEqual(t, got, want)
}

func TestPreprocessSingleLineCommentContentTrimmed(t *testing.T) {
	t.Parallel()

	source := "//   x = 1   \n!!\t  input x\t\n"

	got := mustPreprocess(t, source)
	want := []string{"x = 1", "input x"}

	assertLinesEqual(t, got, want)
}

func TestPreprocessSingleLineCommentSpacesOnlyAfterMarker(t *testing.T) {
	t.Parallel()

	source := "//    \n!!\t\t\n"

	got := mustPreprocess(t, source)
	want := []string{"", ""}

	assertLinesEqual(t, got, want)
}

func TestPreprocessBlockCommentOpenAndCloseSameLineNoPadding(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		source string
		want   []string
	}{
		{name: "slash-star", source: "/*x=1*/\n", want: []string{"x=1"}},
		{name: "bang-star", source: "!*y=2*!\n", want: []string{"y=2"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := mustPreprocess(t, tc.source)
			assertLinesEqual(t, got, tc.want)
		})
	}
}

func TestPreprocessOnlyIgnoredLines(t *testing.T) {
	t.Parallel()

	source := "plain text\n  still plain\nx = 1 // inline marker\n"

	got := mustPreprocess(t, source)
	want := []string{}

	assertLinesEqual(t, got, want)
}

func TestPreprocessBlockCommentOpeningLineWithTrailingContent(t *testing.T) {
	t.Parallel()

	source := "/* opening text\nx = 1\n*/\n"

	got := mustPreprocess(t, source)
	want := []string{"opening text", "x = 1"}

	assertLinesEqual(t, got, want)
}

func TestPreprocessBlockCloseMarkersDoNotCrossMatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		source string
		want   []string
	}{
		{
			name: "slash-open-does-not-close-with-wbang",
			source: "/*\n" +
				"a\n" +
				"*!\n" +
				"b\n" +
				"*/\n",
			want: []string{"a", "*!", "b"},
		},
		{
			name: "bang-open-does-not-close-with-slashstar",
			source: "!*\n" +
				"a\n" +
				"*/\n" +
				"b\n" +
				"*!\n",
			want: []string{"a", "*/", "b"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := mustPreprocess(t, tc.source)
			assertLinesEqual(t, got, tc.want)
		})
	}
}

func TestPreprocessBlockCommentMarkerMidLineIgnored(t *testing.T) {
	t.Parallel()

	source := "ignored /* x = 1 */\n// y = 2\n"

	got := mustPreprocess(t, source)
	want := []string{"y = 2"}

	assertLinesEqual(t, got, want)
}

func TestPreprocessWhitespaceOnlyLineBetweenExecCommentsIgnored(t *testing.T) {
	t.Parallel()

	source := "// x = 1\n   \n// y = 2\n"

	got := mustPreprocess(t, source)
	want := []string{"x = 1", "y = 2"}

	assertLinesEqual(t, got, want)
}

func TestPreprocessSingleLineCommentWithoutTrailingNewline(t *testing.T) {
	t.Parallel()

	source := "// x = 1"

	got := mustPreprocess(t, source)
	want := []string{"x = 1"}

	assertLinesEqual(t, got, want)
}

func TestPreprocessEmptyInput(t *testing.T) {
	t.Parallel()

	got := mustPreprocess(t, "")
	want := []string{}

	assertLinesEqual(t, got, want)
}

func assertLinesEqual(t *testing.T, got, want []string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("lines mismatch\n got: %#v\nwant: %#v", got, want)
	}
}

func mustPreprocess(t *testing.T, source string) []string {
	t.Helper()
	got, err := Preprocess(source)
	if err != nil {
		t.Fatalf("Preprocess() error: %v", err)
	}
	return got
}
