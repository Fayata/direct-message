package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/jung-kurt/gofpdf"
	"github.com/xuri/excelize/v2"
	"phising/models"
)

func writeExcelDynamic(w http.ResponseWriter, fields []models.FormField, submissions []models.Submission) {
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
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", `attachment; filename="data-user-`+time.Now().Format("20060102-150405")+`.xlsx"`)
	if err := f.Write(w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func writePDFDynamic(w http.ResponseWriter, fields []models.FormField, submissions []models.Submission) {
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
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", `attachment; filename="data-user-`+time.Now().Format("20060102-150405")+`.pdf"`)
	if err := pdf.Output(w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max]) + "..."
}
