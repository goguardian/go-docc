package docc

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrNotSupportFormat = errors.New("the file is not supported")
	xmlParagraph        = "p"
	xmlTab              = "t"
)

type FileReader struct {
	fileName string
	xml      io.ReadCloser
	decoder  *xml.Decoder
}

type Reader struct {
	docxPath    string
	fromDoc     bool
	docx        *zip.ReadCloser
	fileReaders []FileReader
}

func getFileReaders(zipReader *zip.Reader) ([]FileReader, error) {
	var fileReaders []FileReader
	for _, file := range zipReader.File {
		if file.Name == "word/document.xml" ||
			strings.Contains(file.Name, "header") ||
			strings.Contains(file.Name, "footer") ||
			strings.Contains(file.Name, "footnotes") {
			openedFile, err := zipReader.Open(file.Name)
			if err != nil {
				return nil, err
			}

			fileReaders = append(fileReaders, FileReader{
				fileName: file.Name,
				xml:      openedFile,
				decoder:  xml.NewDecoder(openedFile),
			})
		}
	}
	return fileReaders, nil
}

// NewReader generetes a Reader struct.
// After reading, the Reader struct shall be Close().
func NewReaderFromBytes(byteValues []byte) (*Reader, error) {
	reader := new(Reader)

	byteReader := bytes.NewReader(byteValues)
	zipReader, err := zip.NewReader(byteReader, int64(len(byteValues)))
	if err != nil {
		return nil, err
	}

	fileReaders, err := getFileReaders(zipReader)
	if err != nil {
		return nil, err
	}
	reader.fileReaders = fileReaders

	return reader, nil
}

// NewReader generetes a Reader struct.
// After reading, the Reader struct shall be Close().
func NewReader(docxPath string) (*Reader, error) {
	reader := new(Reader)
	reader.docxPath = docxPath
	ext := strings.ToLower(filepath.Ext(docxPath))
	if ext != ".docx" {
		return nil, ErrNotSupportFormat
	}

	zipReadCloser, err := zip.OpenReader(reader.docxPath)
	if err != nil {
		return nil, err
	}

	fileReaders, err := getFileReaders(&zipReadCloser.Reader)
	if err != nil {
		return nil, err
	}
	reader.fileReaders = fileReaders

	reader.docx = zipReadCloser

	return reader, nil
}

func (r *Reader) read(decoder *xml.Decoder) (string, error) {
	err := seekNextTag(decoder, xmlParagraph)
	if err != nil {
		return "", err
	}
	paragraph, err := seekParagraph(decoder)
	if err != nil {
		return "", err
	}
	return paragraph, nil
}

func (r *Reader) readSingleFile(decoder *xml.Decoder) (string, error) {
	var content strings.Builder
	for {
		paragraph, err := r.read(decoder)
		if err == io.EOF {
			return content.String(), nil
		} else if err != nil {
			return "", err
		}
		content.WriteString(paragraph)
		content.WriteString(" ")
	}
}

// ReadAllFiles reads all header, footer, footnote and content related files in
// the zip archive and returns it's raw contents
func (r *Reader) ReadAllFiles() (headerValue, contentValue, footerValue, footnotesValue string, err error) {
	var header, footer, footnotes, content strings.Builder
	for _, fileReader := range r.fileReaders {
		fileContent, err := r.readSingleFile(fileReader.decoder)
		if err != nil {
			return "", "", "", "", err
		}

		if strings.Contains(fileReader.fileName, "header") {
			header.WriteString(fileContent)
		} else if strings.Contains(fileReader.fileName, "footer") {
			footer.WriteString(fileContent)
		} else if strings.Contains(fileReader.fileName, "footnotes") {
			footnotes.WriteString(fileContent)
		} else {
			content.WriteString(fileContent)
		}
	}

	return trimmedString(header.String()),
		trimmedString(content.String()),
		trimmedString(footer.String()),
		trimmedString(footnotes.String()),
		nil
}

func trimmedString(input string) string {
	return strings.TrimSpace(input)
}

func (r *Reader) Close() error {
	for _, fileReader := range r.fileReaders {
		fileReader.xml.Close()
	}
	if r.docx != nil {
		r.docx.Close()
	}
	if r.fromDoc {
		os.Remove(r.docxPath)
	}
	return nil
}

func seekText(decoder *xml.Decoder) (string, error) {
	for {
		token, err := decoder.Token()
		if err != nil {
			return "", err
		}
		switch tokenType := token.(type) {
		case xml.CharData:
			return string(tokenType), nil
		case xml.EndElement:
			return "", nil
		}
	}
}

func seekParagraph(decoder *xml.Decoder) (string, error) {
	var paragraph strings.Builder
	for {
		token, err := decoder.Token()
		if err != nil {
			return "", err
		}
		switch tokenType := token.(type) {
		case xml.EndElement:
			if tokenType.Name.Local == xmlParagraph {
				return paragraph.String(), nil
			}
		case xml.StartElement:
			if tokenType.Name.Local == xmlTab {
				text, err := seekText(decoder)
				if err != nil {
					return "", err
				}
				paragraph.WriteString(text)
			}
		}
	}
}

func seekNextTag(decoder *xml.Decoder, tag string) error {
	for {
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		tokenTag, ok := token.(xml.StartElement)
		if !ok {
			continue
		}
		if tokenTag.Name.Local != tag {
			continue
		}
		break
	}
	return nil
}
