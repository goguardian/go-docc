package docc

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

var expectedContent = map[string][]string{
	"./testdata/test.docx":               {"", "Title Subtitle Here is a first row. Here is a second row.", "", ""},
	"./testdata/test_header_footer.docx": {"test header", "Title Subtitle Here is a first row. Here is a second row.", "test footer", ""},
}

func TestReaderReadAll(t *testing.T) {
	for fileName, expectContent := range expectedContent {
		filePath := filepath.Clean(fileName)
		r, err := NewReader(filePath)
		if err != nil {
			panic(err)
		}

		defer r.Close()
		header, content, footer, footnotes, err := r.ReadAllFiles()
		if err != nil {
			panic(err)
		}

		if !reflect.DeepEqual(expectContent[0], header) {
			t.Errorf("want header %v, got %v for fileName %v", expectContent[0], header, fileName)
		}
		if !reflect.DeepEqual(expectContent[1], content) {
			t.Errorf("want content %v, got %v for fileName %v", expectContent[1], content, fileName)
		}
		if !reflect.DeepEqual(expectContent[2], footer) {
			t.Errorf("want footer %v, got %v for fileName %v", expectContent[2], footer, fileName)
		}
		if !reflect.DeepEqual(expectContent[3], footnotes) {
			t.Errorf("want footnotes %v, got %v for fileName %v", expectContent[3], footnotes, fileName)
		}
	}
}

func TestReaderFromBytesReadAll(t *testing.T) {
	for fileName, expectContent := range expectedContent {
		filePath := filepath.Clean(fileName)
		fileBytes, err := os.ReadFile(filePath)
		if err != nil {
			panic(err)
		}

		r, err := NewReaderFromBytes(fileBytes)
		if err != nil {
			panic(err)
		}
		defer r.Close()
		header, content, footer, footnotes, err := r.ReadAllFiles()
		if err != nil {
			panic(err)
		}

		if !reflect.DeepEqual(expectContent[0], header) {
			t.Errorf("want header %v, got %v for fileName %v", expectContent[0], header, fileName)
		}
		if !reflect.DeepEqual(expectContent[1], content) {
			t.Errorf("want content %v, got %v for fileName %v", expectContent[1], content, fileName)
		}
		if !reflect.DeepEqual(expectContent[2], footer) {
			t.Errorf("want footer %v, got %v for fileName %v", expectContent[2], footer, fileName)
		}
		if !reflect.DeepEqual(expectContent[3], footnotes) {
			t.Errorf("want footnotes %v, got %v for fileName %v", expectContent[3], footnotes, fileName)
		}
	}
}
