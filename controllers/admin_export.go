package controllers

import (
	"archive/zip"
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"
	"github.com/xuri/excelize/v2"
	"phising/models"
)

func buildExcelDynamic(fields []models.FormField, submissions []models.Submission) ([]byte, error) {
	f := excelize.NewFile()
	sheet := "Sheet1"
	colIdx := 1
	cell, _ := excelize.CoordinatesToCellName(colIdx, 1)
	_ = f.SetCellValue(sheet, cell, "No")
	colIdx++
	for _, h := range fields {
		cell, _ = excelize.CoordinatesToCellName(colIdx, 1)
		_ = f.SetCellValue(sheet, cell, h.Label)
		colIdx++
	}
	cell, _ = excelize.CoordinatesToCellName(colIdx, 1)
	_ = f.SetCellValue(sheet, cell, "Waktu Submit")
	for i, s := range submissions {
		row := i + 2
		c, _ := excelize.CoordinatesToCellName(1, row)
		_ = f.SetCellValue(sheet, c, i+1)
		for j, field := range fields {
			c, _ = excelize.CoordinatesToCellName(j+2, row)
			_ = f.SetCellValue(sheet, c, s.Data[field.Name])
		}
		c, _ = excelize.CoordinatesToCellName(len(fields)+2, row)
		createdStr := ""
		if !s.CreatedAt.IsZero() {
			createdStr = s.CreatedAt.Format("02 Jan 2006 15:04")
		}
		_ = f.SetCellValue(sheet, c, createdStr)
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func buildPDFDynamic(fields []models.FormField, submissions []models.Submission) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddPage()
	pdf.SetFont("Helvetica", "B", 12)
	pdf.CellFormat(0, 8, "Data User - Form", "", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 8)
	pdf.Ln(4)
	nCols := len(fields) + 2
	colW := 190.0 / float64(nCols)
	if colW > 40 {
		colW = 40
	}
	headers := []string{"No"}
	for _, h := range fields {
		headers = append(headers, truncate(h.Label, 12))
	}
	headers = append(headers, "Waktu")
	for _, h := range headers {
		pdf.CellFormat(colW, 6, h, "1", 0, "L", true, 0, "")
	}
	pdf.Ln(-1)
	for i, s := range submissions {
		pdf.CellFormat(colW, 5, strconv.Itoa(i+1), "1", 0, "L", false, 0, "")
		for _, field := range fields {
			pdf.CellFormat(colW, 5, truncate(s.Data[field.Name], 15), "1", 0, "L", false, 0, "")
		}
		createdStr := ""
		if !s.CreatedAt.IsZero() {
			createdStr = s.CreatedAt.Format("02/01/06 15:04")
		}
		pdf.CellFormat(colW, 5, createdStr, "1", 0, "L", false, 0, "")
		pdf.Ln(-1)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func writeAttachment(w http.ResponseWriter, contentType, filename string, body []byte) {
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	_, _ = io.Copy(w, bytes.NewReader(body))
}

func randomExportKey() (string, error) {
	b := make([]byte, 24) // 32 chars-ish after base64url
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func buildKeyBundleZip() (zipBody []byte, key string, err error) {
	key, err = randomExportKey()
	if err != nil {
		return nil, "", err
	}
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	meta := map[string]string{
		"format":      "phising-export-key-v1",
		"generatedAt": time.Now().Format(time.RFC3339),
		"note":        "Jaga file ini. Key dipakai untuk encrypted export.",
	}
	metaBytes, _ := json.MarshalIndent(meta, "", "  ")

	f1, err := zw.Create("README.txt")
	if err != nil {
		return nil, "", err
	}
	_, _ = f1.Write([]byte("File key untuk membuka export terenkripsi.\nSimpan aman, jangan dibagikan.\n"))

	f2, err := zw.Create("key.txt")
	if err != nil {
		return nil, "", err
	}
	_, _ = f2.Write([]byte(key))

	f3, err := zw.Create("meta.json")
	if err != nil {
		return nil, "", err
	}
	_, _ = f3.Write(metaBytes)

	if err := zw.Close(); err != nil {
		return nil, "", err
	}
	return buf.Bytes(), key, nil
}

func readKeyFromZip(zipData []byte) (string, error) {
	zr, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return "", fmt.Errorf("invalid key zip: %w", err)
	}
	for _, f := range zr.File {
		if !strings.EqualFold(f.Name, "key.txt") {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return "", err
		}
		defer rc.Close()
		b, err := io.ReadAll(rc)
		if err != nil {
			return "", err
		}
		key := strings.TrimSpace(string(b))
		if key == "" {
			return "", errors.New("key.txt kosong")
		}
		return key, nil
	}
	return "", errors.New("key.txt tidak ditemukan di zip")
}

func encryptWithPython(in []byte, passphrase string) ([]byte, error) {
	script := filepath.Join("tools", "encrypt_export.py")
	if _, err := os.Stat(script); err != nil {
		return nil, fmt.Errorf("encrypt script not found: %w", err)
	}
	if passphrase == "" {
		return nil, errors.New("empty passphrase")
	}

	inFile, err := os.CreateTemp("", "phising-export-*.bin")
	if err != nil {
		return nil, err
	}
	defer os.Remove(inFile.Name())
	if _, err := inFile.Write(in); err != nil {
		_ = inFile.Close()
		return nil, err
	}
	_ = inFile.Close()

	outFile, err := os.CreateTemp("", "phising-export-*.enc")
	if err != nil {
		return nil, err
	}
	outName := outFile.Name()
	_ = outFile.Close()
	defer os.Remove(outName)

	cmd := exec.Command("python", script, "--mode", "encrypt", "--in", inFile.Name(), "--out", outName, "--passphrase", passphrase)
	// Run from repo root so relative "tools/..." works.
	cmd.Dir = "."
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("python encrypt failed: %w (%s)", err, strings.TrimSpace(string(out)))
	}
	enc, err := os.ReadFile(outName)
	if err != nil {
		return nil, err
	}
	return enc, nil
}

func decryptWithPython(in []byte, passphrase string) ([]byte, error) {
	script := filepath.Join("tools", "encrypt_export.py")
	if _, err := os.Stat(script); err != nil {
		return nil, fmt.Errorf("decrypt script not found: %w", err)
	}
	if passphrase == "" {
		return nil, errors.New("empty passphrase")
	}

	inFile, err := os.CreateTemp("", "phising-encrypted-*.enc")
	if err != nil {
		return nil, err
	}
	defer os.Remove(inFile.Name())
	if _, err := inFile.Write(in); err != nil {
		_ = inFile.Close()
		return nil, err
	}
	_ = inFile.Close()

	outFile, err := os.CreateTemp("", "phising-decrypted-*.bin")
	if err != nil {
		return nil, err
	}
	outName := outFile.Name()
	_ = outFile.Close()
	defer os.Remove(outName)

	cmd := exec.Command("python", script, "--mode", "decrypt", "--in", inFile.Name(), "--out", outName, "--passphrase", passphrase)
	cmd.Dir = "."
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("python decrypt failed (key salah / file rusak): %w (%s)", err, strings.TrimSpace(string(out)))
	}
	plain, err := os.ReadFile(outName)
	if err != nil {
		return nil, err
	}
	return plain, nil
}

func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max]) + "..."
}
