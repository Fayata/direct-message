package main

import (
	"fmt"
	"log"
	"net/http"

	"phising/controllers"
)

func main() {
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", controllers.IndexHandler)
	http.HandleFunc("/api/templates", controllers.TemplatesHandler)
	http.HandleFunc("/api/templates/save", controllers.SaveTemplateHandler)
	http.HandleFunc("/api/send", controllers.SendMessageHandler)
	http.HandleFunc("/form", controllers.FormHandler)

	http.HandleFunc("/admin/login", controllers.AdminLoginHandler)
	http.HandleFunc("/admin/logout", controllers.AdminLogoutHandler)
	http.HandleFunc("/admin/delete", controllers.RequireAdmin(controllers.AdminDeleteSubmissionHandler))
	http.HandleFunc("/admin/bulk-delete", controllers.RequireAdmin(controllers.AdminBulkDeleteSubmissionsHandler))
	http.HandleFunc("/admin/settings", controllers.RequireAdmin(controllers.AdminSettingsHandler))
	http.HandleFunc("/admin/form/", controllers.RequireAdmin(controllers.AdminFormHandler))
	http.HandleFunc("/admin/export", controllers.RequireAdmin(controllers.AdminExportHandler))
	http.HandleFunc("/admin", controllers.RequireAdmin(controllers.AdminDashboardHandler))
	http.HandleFunc("/admin/", controllers.RequireAdmin(controllers.AdminDashboardHandler))

	fmt.Println("Server berjalan di https://localhost:8080")
	log.Fatal(http.ListenAndServeTLS(":8080", "cert.pem", "key.pem", nil))
}
