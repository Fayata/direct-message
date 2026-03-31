package controllers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"phising/models"
)

var (
	indexTmpl = template.Must(template.ParseFiles(filepath.Join("views", "index.html")))
	formTplFuncs = template.FuncMap{
		"splitOptions": func(s string) []string {
			if s == "" {
				return nil
			}
			var out []string
			for _, v := range strings.Split(s, ",") {
				out = append(out, strings.TrimSpace(v))
			}
			return out
		},
		"formValue": func(values map[string]string, name string) string {
			if values == nil {
				return ""
			}
			return values[name]
		},
		"fieldKind": func(ft string) string {
			switch strings.ToLower(strings.TrimSpace(ft)) {
			case "textarea":
				return "textarea"
			case "select":
				return "select"
			default:
				return "input"
			}
		},
		"inputType": func(ft string) string {
			switch strings.ToLower(strings.TrimSpace(ft)) {
			case "email", "number", "date", "tel":
				return strings.ToLower(strings.TrimSpace(ft))
			default:
				return "text"
			}
		},
	}
	formTmpl  = template.Must(template.New("form.html").Funcs(formTplFuncs).ParseFiles(filepath.Join("views", "form.html")))
	form2Tmpl = template.Must(template.New("form2.html").Funcs(formTplFuncs).ParseFiles(filepath.Join("views", "form2.html")))
)

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := indexTmpl.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func TemplatesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(models.Templates); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func SaveTemplateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var newTemplate struct {
		Name    string `json:"name"`
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&newTemplate); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	templateModel := models.MessageTemplate{
		ID:      string(rune(len(models.Templates) + 1)),
		Name:    newTemplate.Name,
		Content: newTemplate.Content,
	}

	models.Templates = append(models.Templates, templateModel)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(templateModel); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func SendMessageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var form models.MessageForm
	if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Mengirim pesan ke %s: %s", form.PhoneNumber, form.Message)

	response := map[string]string{
		"status":  "success",
		"message": "Pesan siap dikirim",
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func formData(settings map[string]string, success bool, errMsg string) map[string]any {
	data := map[string]any{
		"Success": success,
		"Error":   errMsg,
		"FormTitle":           settings["form_title"],
		"FormSubtitle":        settings["form_subtitle"],
		"FormSuccessMessage":  settings["form_success_message"],
		"FormHelpText":        settings["form_help_text"],
	}
	if data["FormTitle"] == "" {
		data["FormTitle"] = "Form Data Diri"
	}
	if data["FormSubtitle"] == "" {
		data["FormSubtitle"] = "Silakan lengkapi data Anda."
	}
	if data["FormSuccessMessage"] == "" {
		data["FormSuccessMessage"] = "Terima kasih, data Anda berhasil disimpan. Anda dapat menutup halaman ini."
	}
	if data["FormHelpText"] == "" {
		data["FormHelpText"] = "Pastikan data yang Anda isi sudah benar sebelum menekan tombol submit."
	}
	return data
}

func splitFieldsWizard(fields []models.FormField) (part1, part2 []models.FormField) {
	if len(fields) == 0 {
		return nil, nil
	}
	mid := (len(fields) + 1) / 2
	return fields[:mid], fields[mid:]
}

// pairFieldsForLayout mengelompokkan field menjadi baris 2 kolom (textarea = baris penuh).
func pairFieldsForLayout(fields []models.FormField) [][]models.FormField {
	var rows [][]models.FormField
	i := 0
	for i < len(fields) {
		f := fields[i]
		if strings.EqualFold(strings.TrimSpace(f.FieldType), "textarea") {
			rows = append(rows, []models.FormField{f})
			i++
			continue
		}
		if i+1 < len(fields) {
			n2 := fields[i+1]
			if strings.EqualFold(strings.TrimSpace(n2.FieldType), "textarea") {
				rows = append(rows, []models.FormField{f})
				i++
				continue
			}
			rows = append(rows, []models.FormField{f, n2})
			i += 2
			continue
		}
		rows = append(rows, []models.FormField{f})
		i++
	}
	return rows
}

func formWizardLayout(fields []models.FormField) (rows1, rows2 [][]models.FormField, hasPart2 bool) {
	p1, p2 := splitFieldsWizard(fields)
	rows1 = pairFieldsForLayout(p1)
	rows2 = pairFieldsForLayout(p2)
	hasPart2 = len(p2) > 0
	return rows1, rows2, hasPart2
}

func buildFormPageData(settings map[string]string, success bool, errMsg string, fields []models.FormField, values map[string]string) map[string]any {
	data := formData(settings, success, errMsg)
	data["Fields"] = fields
	r1, r2, has := formWizardLayout(fields)
	data["FieldRows1"] = r1
	data["FieldRows2"] = r2
	data["HasPart2"] = has
	if values == nil {
		values = map[string]string{}
	}
	data["Values"] = values
	return data
}

func FormHandler(w http.ResponseWriter, r *http.Request) {
	settings, _ := models.GetFormSettings(r.Context())
	if settings == nil {
		settings = make(map[string]string)
	}
	fields, _ := models.ListFormFields(r.Context())
	if fields == nil {
		fields = []models.FormField{}
	}

	switch r.Method {
	case http.MethodGet:
		data := buildFormPageData(settings, false, "", fields, nil)
		data["FormPostPath"] = "/form"
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := formTmpl.Execute(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		wantsJSON := strings.Contains(r.Header.Get("Accept"), "application/json")
		dataMap := make(map[string]string)
		for _, f := range fields {
			val := r.FormValue(f.Name)
			dataMap[f.Name] = val
			if f.Required && val == "" {
				if wantsJSON {
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(map[string]string{
						"status":  "error",
						"message": "Semua field wajib diisi.",
					})
					return
				}
				data := buildFormPageData(settings, false, "Semua field wajib diisi.", fields, dataMap)
				data["FormPostPath"] = "/form"
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.Header().Set("X-Form-Status", "error")
				_ = formTmpl.Execute(w, data)
				return
			}
		}
		id, err := models.SaveSubmission(r.Context(), dataMap)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		ref := fmt.Sprintf("AP-%d", id)
		if wantsJSON {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.Header().Set("X-Form-Status", "ok")
			w.Header().Set("X-Form-Ref", ref)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"status": "ok",
				"ref":    ref,
			})
			return
		}
		data := buildFormPageData(settings, true, "", fields, map[string]string{})
		data["FormPostPath"] = "/form"
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("X-Form-Status", "ok")
		w.Header().Set("X-Form-Ref", ref)
		if err := formTmpl.Execute(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func Form2Handler(w http.ResponseWriter, r *http.Request) {
	settings, _ := models.GetFormSettings(r.Context())
	if settings == nil {
		settings = make(map[string]string)
	}
	fields, _ := models.ListFormFields(r.Context())
	if fields == nil {
		fields = []models.FormField{}
	}

	switch r.Method {
	case http.MethodGet:
		data := buildFormPageData(settings, false, "", fields, nil)
		data["FormPostPath"] = "/form2"
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := form2Tmpl.Execute(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		wantsJSON := strings.Contains(r.Header.Get("Accept"), "application/json")
		dataMap := make(map[string]string)
		for _, f := range fields {
			val := r.FormValue(f.Name)
			dataMap[f.Name] = val
			if f.Required && val == "" {
				if wantsJSON {
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(map[string]string{
						"status":  "error",
						"message": "Semua field wajib diisi.",
					})
					return
				}
				data := buildFormPageData(settings, false, "Semua field wajib diisi.", fields, dataMap)
				data["FormPostPath"] = "/form2"
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.Header().Set("X-Form-Status", "error")
				_ = form2Tmpl.Execute(w, data)
				return
			}
		}
		id, err := models.SaveSubmission(r.Context(), dataMap)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		ref := fmt.Sprintf("AP-%d", id)
		if wantsJSON {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.Header().Set("X-Form-Status", "ok")
			w.Header().Set("X-Form-Ref", ref)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"status": "ok",
				"ref":    ref,
			})
			return
		}
		data := buildFormPageData(settings, true, "", fields, map[string]string{})
		data["FormPostPath"] = "/form2"
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("X-Form-Status", "ok")
		w.Header().Set("X-Form-Ref", ref)
		if err := form2Tmpl.Execute(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

