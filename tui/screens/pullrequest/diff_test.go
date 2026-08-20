package pullrequest

import "testing"

const sampleDiff = `diff --git a/src/cart.go b/src/cart.go
index 1111111..2222222 100644
--- a/src/cart.go
+++ b/src/cart.go
@@ -1,4 +1,5 @@
 package cart
+import "fmt"

-func Old() {}
+func New() { fmt.Println("hi") }
\ No newline at end of file
diff --git a/gone.txt b/gone.txt
deleted file mode 100644
index 3333333..0000000
--- a/gone.txt
+++ /dev/null
@@ -1,1 +0,0 @@
-goodbye
diff --git a/old-name.txt b/new-name.txt
similarity index 90%
rename from old-name.txt
rename to new-name.txt
--- a/old-name.txt
+++ b/new-name.txt
@@ -1 +1 @@
-was
+is
`

func TestParseDiff_ShouldSplitFilesAndHunks(t *testing.T) {
	// Act
	files := ParseDiff(sampleDiff)

	// Assert
	if len(files) != 3 {
		t.Fatalf("expected 3 files, got %d: %+v", len(files), files)
	}
	if files[0].Path != "src/cart.go" {
		t.Fatalf("expected the +++ side's path, got %q", files[0].Path)
	}
	// Header, context, add, empty context, remove, add. The no-newline marker
	// and the index/--- metadata must not leak in.
	kinds := ""
	for _, hunk := range files[0].Hunks {
		kinds += string(hunk.Kind)
	}
	if kinds != "@ + -+" {
		t.Fatalf("unexpected hunk kinds %q: %+v", kinds, files[0].Hunks)
	}
	if files[0].Hunks[2].Text != `import "fmt"` {
		t.Fatalf("expected the marker stripped from the text, got %q", files[0].Hunks[2].Text)
	}
}

func TestParseDiff_ShouldKeepADeletedFilesName(t *testing.T) {
	// Act: the deleted file's +++ side is /dev/null, so its name has to come
	// from the diff --git header.
	files := ParseDiff(sampleDiff)

	// Assert
	if files[1].Path != "gone.txt" {
		t.Fatalf("expected the deleted file's old name, got %q", files[1].Path)
	}
}

func TestParseDiff_ShouldFollowARename(t *testing.T) {
	// Act
	files := ParseDiff(sampleDiff)

	// Assert
	if files[2].Path != "new-name.txt" {
		t.Fatalf("expected the renamed file's new name, got %q", files[2].Path)
	}
}

func TestParseDiff_ShouldReturnNothingForAnEmptyDiff(t *testing.T) {
	if files := ParseDiff(""); len(files) != 0 {
		t.Fatalf("expected no files for an empty diff, got %+v", files)
	}
}

func TestTwoColumnRows_ShouldAlignHeadersContextAndUnequalReplacements(t *testing.T) {
	// Arrange
	hunks := []DiffHunk{
		{Kind: '@', Text: "@@ -1 +1 @@"},
		{Kind: '-', Text: "old one"},
		{Kind: '-', Text: "old two"},
		{Kind: '+', Text: "new one"},
		{Kind: ' ', Text: "same"},
		{Kind: '+', Text: "new three"},
	}

	// Act
	rows := twoColumnRows(hunks)

	// Assert
	if len(rows) != 5 || !rows[0].header || rows[0].text != "@@ -1 +1 @@" {
		t.Fatalf("expected a full-width hunk header, got %+v", rows)
	}
	if rows[1].left == nil || rows[1].right == nil || rows[1].left.Text != "old one" ||
		rows[1].right.Text != "new one" {
		t.Fatalf("expected the first replacement pair, got %+v", rows[1])
	}
	if rows[2].left == nil || rows[2].right != nil {
		t.Fatalf("expected an unmatched removal opposite a blank, got %+v", rows[2])
	}
	if rows[3].left == nil || rows[3].right == nil ||
		rows[3].left.Text != "same" || rows[3].right.Text != "same" {
		t.Fatalf("expected duplicated context, got %+v", rows[3])
	}
	if rows[4].left != nil || rows[4].right == nil || rows[4].right.Text != "new three" {
		t.Fatalf("expected an unmatched addition opposite a blank, got %+v", rows[4])
	}
}
