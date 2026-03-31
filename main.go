package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

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
	http.HandleFunc("/form2", controllers.Form2Handler)

	http.HandleFunc("/admin/login", controllers.AdminLoginHandler)
	http.HandleFunc("/admin/logout", controllers.AdminLogoutHandler)
	http.HandleFunc("/admin/delete", controllers.RequireAdmin(controllers.AdminDeleteSubmissionHandler))
	http.HandleFunc("/admin/bulk-delete", controllers.RequireAdmin(controllers.AdminBulkDeleteSubmissionsHandler))
	http.HandleFunc("/admin/settings", controllers.RequireAdmin(controllers.AdminSettingsHandler))
	http.HandleFunc("/admin/form/", controllers.RequireAdmin(controllers.AdminFormHandler))
	http.HandleFunc("/admin/form2/", controllers.RequireAdmin(controllers.AdminForm2Handler))
	http.HandleFunc("/admin/form2", controllers.RequireAdmin(controllers.AdminForm2Handler))
	http.HandleFunc("/admin/export", controllers.RequireAdmin(controllers.AdminExportHandler))
	http.HandleFunc("/admin", controllers.RequireAdmin(controllers.AdminDashboardHandler))
	http.HandleFunc("/admin/", controllers.RequireAdmin(controllers.AdminDashboardHandler))

	fmt.Println("Server berjalan di https://localhost:8080")

	certPath := "cert.pem"
	keyPath := "key.pem"
	_, certErr := os.Stat(certPath)
	_, keyErr := os.Stat(keyPath)
	if os.IsNotExist(certErr) || os.IsNotExist(keyErr) {
		log.Printf("TLS cert/key belum ada (%v / %v). Membuat ulang...", certPath, keyPath)
		generateCertificate()
	}

	log.Fatal(http.ListenAndServeTLS(":8080", certPath, keyPath, nil))
}
