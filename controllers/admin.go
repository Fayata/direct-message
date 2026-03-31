package controllers

import (
	"html/template"
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
		"formatTime": func(t time.Time) string {
			if t.IsZero() {
				return "-"
			}
			return t.Format("02 Jan 2006 15:04")
		},
	}).ParseFiles(filepath.Join("views", "admin", "dashboard.html")))
	adminFormTmpl = template.Must(template.New("form.html").Funcs(template.FuncMap{
		"add": func(a, b int) int { return a + b },
	}).ParseFiles(filepath.Join("views", "admin", "form.html")))
	adminSettings = template.Must(template.ParseFiles(filepath.Join("views", "admin", "settings.html")))
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

	fields, err := models.ListFormFields(r.Context())
	if err != nil {
		fields = []models.FormField{}
	}

	pageSize := 50
	page := 1
	if p, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil && p > 0 {
		page = p
	}
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	fieldKey := strings.TrimSpace(r.URL.Query().Get("field"))

	// Total rows untuk pagination (tanpa filter q)
	total, err := models.CountSubmissions(r.Context())
	if err != nil {
		log.Printf("CountSubmissions: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	maxPage := (total + pageSize - 1) / pageSize
	if maxPage < 1 {
		maxPage = 1
	}
	if page > maxPage {
		page = maxPage
	}
	offset := (page - 1) * pageSize

	submissions, err := models.ListSubmissionsPage(r.Context(), pageSize, offset)
	if err != nil {
		log.Printf("ListSubmissionsPage: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := adminDashboard.Execute(w, map[string]any{
		"Fields":      fields,
		"Submissions": submissions,
		"Page":        page,
		"MaxPage":     maxPage,
		"PageSize":    pageSize,
		"Total":       total,
		"Q":           q,
		"FieldKey":    fieldKey,
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
	if err := models.DeleteSubmission(r.Context(), id); err != nil {
		log.Printf("DeleteSubmission: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin", http.StatusFound)
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
	ids := r.Form["ids"]
	for _, idStr := range ids {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil || id <= 0 {
			continue
		}
		_ = models.DeleteSubmission(r.Context(), id)
	}
	http.Redirect(w, r, "/admin", http.StatusFound)
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
			"Settings": settings,
			"Saved":    r.URL.Query().Get("saved") == "1",
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
	fields, _ := models.ListFormFields(r.Context())
	submissions, err := models.ListSubmissions(r.Context())
	if err != nil {
		log.Printf("ListSubmissions: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	switch format {
	case "xlsx", "excel":
		writeExcelDynamic(w, fields, submissions)
	case "pdf":
		writePDFDynamic(w, fields, submissions)
	default:
		http.Error(w, "format tidak valid (gunakan format=xlsx atau format=pdf)", http.StatusBadRequest)
	}
}
