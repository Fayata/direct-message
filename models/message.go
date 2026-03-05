package models

type MessageTemplate struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Content string `json:"content"`
}

type MessageForm struct {
	PhoneNumber string `json:"phone_number"`
	Message     string `json:"message"`
	TemplateID  string `json:"template_id"`
}

var Templates []MessageTemplate

func init() {
	Templates = []MessageTemplate{
		{ID: "1", Name: "Salam", Content: "Halo {nama}, semoga harimu menyenangkan!"},
		{ID: "2", Name: "Promosi", Content: "Hai {nama}, ada promo spesial untuk kamu! Dapatkan diskon hingga {diskon}%."},
		{ID: "3", Name: "Reminder", Content: "Halo {nama}, ini adalah pengingat untuk {kegiatan} pada tanggal {tanggal}."},
		{ID: "4", Name: "Custom", Content: ""},
	}
}

