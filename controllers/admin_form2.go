package controllers

import "net/http"

func AdminForm2Handler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/admin/form", http.StatusFound)
}

