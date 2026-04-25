package models

import (
	"reflect"
	"testing"
)

func TestExtractCommandTemplateVariables(t *testing.T) {
	content := `ssh $$$host$$$ -p ***port*** && export USER=$$$user_name$$$ && echo "***port***"`

	got := ExtractCommandTemplateVariables(content)
	want := []string{"host", "port", "user_name"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected variables: got %v want %v", got, want)
	}
}

func TestReplaceCommandTemplateVariables(t *testing.T) {
	content := `scp $$$file$$$ ***user***@$$$host$$$:/tmp && echo ***missing***`

	rendered, missing := ReplaceCommandTemplateVariables(content, map[string]string{
		"file": "build.tar.gz",
		"user": "ubuntu",
		"host": "10.0.0.8",
	})

	wantRendered := `scp build.tar.gz ubuntu@10.0.0.8:/tmp && echo ***missing***`
	if rendered != wantRendered {
		t.Fatalf("unexpected rendered content: got %q want %q", rendered, wantRendered)
	}

	wantMissing := []string{"missing"}
	if !reflect.DeepEqual(missing, wantMissing) {
		t.Fatalf("unexpected missing variables: got %v want %v", missing, wantMissing)
	}
}
