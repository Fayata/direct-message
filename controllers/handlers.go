package controllers

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"phising/models"
)

var (
	indexTmpl = template.Must(template.ParseFiles(filepath.Join("views", "index.html")))
	formTmpl  = template.Must(template.New("form.html").Funcs(template.FuncMap{
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
	}).ParseFiles(filepath.Join("views", "form.html")))
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
		data := formData(settings, false, "")
		data["Fields"] = fields
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := formTmpl.Execute(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		dataMap := make(map[string]string)
		for _, f := range fields {
			val := r.FormValue(f.Name)
			dataMap[f.Name] = val
			if f.Required && val == "" {
				data := formData(settings, false, "Semua field wajib diisi.")
				data["Fields"] = fields
				data["Values"] = dataMap
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				_ = formTmpl.Execute(w, data)
				return
			}
		}
		if _, err := models.SaveSubmission(r.Context(), dataMap); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data := formData(settings, true, "")
		data["Fields"] = fields
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := formTmpl.Execute(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

