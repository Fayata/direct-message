package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
)

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

var templates []MessageTemplate

func init() {
	// Template pesan default
	templates = []MessageTemplate{
		{ID: "1", Name: "Salam", Content: "Halo {nama}, semoga harimu menyenangkan!"},
		{ID: "2", Name: "Promosi", Content: "Hai {nama}, ada promo spesial untuk kamu! Dapatkan diskon hingga {diskon}%. Info lebih lanjut: {link}"},
		{ID: "3", Name: "Reminder", Content: "Halo {nama}, ini adalah pengingat untuk {kegiatan} pada tanggal {tanggal}."},
		{ID: "4", Name: "Custom", Content: ""},
	}
}

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/api/templates", templatesHandler)
	http.HandleFunc("/api/templates/save", saveTemplateHandler)
	http.HandleFunc("/api/send", sendMessageHandler)

	fmt.Println("Server berjalan di http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	htmlContent := `<!DOCTYPE html>
<html lang="id">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>WhatsApp Message Sender</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            padding: 20px;
        }
        
        .container {
            max-width: 800px;
            margin: 0 auto;
            background: white;
            border-radius: 20px;
            box-shadow: 0 20px 60px rgba(0,0,0,0.3);
            overflow: hidden;
        }
        
        .header {
            background: linear-gradient(135deg, #25D366 0%, #128C7E 100%);
            color: white;
            padding: 30px;
            text-align: center;
        }
        
        .header h1 {
            font-size: 2em;
            margin-bottom: 10px;
        }
        
        .header p {
            opacity: 0.9;
        }
        
        .content {
            padding: 30px;
        }
        
        .form-group {
            margin-bottom: 25px;
        }
        
        label {
            display: block;
            margin-bottom: 8px;
            font-weight: 600;
            color: #333;
        }
        
        input[type="text"],
        input[type="tel"],
        textarea,
        select {
            width: 100%;
            padding: 12px 15px;
            border: 2px solid #e0e0e0;
            border-radius: 10px;
            font-size: 14px;
            transition: all 0.3s;
        }
        
        input[type="text"]:focus,
        input[type="tel"]:focus,
        textarea:focus,
        select:focus {
            outline: none;
            border-color: #25D366;
            box-shadow: 0 0 0 3px rgba(37, 211, 102, 0.1);
        }
        
        textarea {
            min-height: 150px;
            resize: vertical;
            font-family: inherit;
        }
        
        .button-group {
            display: flex;
            gap: 15px;
            margin-top: 30px;
        }
        
        button {
            flex: 1;
            padding: 15px;
            border: none;
            border-radius: 10px;
            font-size: 16px;
            font-weight: 600;
            cursor: pointer;
            transition: all 0.3s;
        }
        
        .btn-primary {
            background: linear-gradient(135deg, #25D366 0%, #128C7E 100%);
            color: white;
        }
        
        .btn-primary:hover {
            transform: translateY(-2px);
            box-shadow: 0 5px 15px rgba(37, 211, 102, 0.3);
        }
        
        .btn-secondary {
            background: #f5f5f5;
            color: #333;
        }
        
        .btn-secondary:hover {
            background: #e0e0e0;
        }
        
        .template-buttons {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
            gap: 10px;
            margin-bottom: 20px;
        }
        
        .template-btn {
            padding: 12px;
            background: #f8f9fa;
            border: 2px solid #e0e0e0;
            border-radius: 8px;
            cursor: pointer;
            transition: all 0.3s;
            font-size: 14px;
        }
        
        .template-btn:hover,
        .template-btn.active {
            background: #25D366;
            color: white;
            border-color: #25D366;
        }
        
        .info-box {
            background: #e3f2fd;
            border-left: 4px solid #2196F3;
            padding: 15px;
            border-radius: 5px;
            margin-bottom: 20px;
        }
        
        .info-box p {
            margin: 5px 0;
            font-size: 14px;
            color: #1976D2;
        }
        
        .template-variables {
            background: #fff3cd;
            border-left: 4px solid #ffc107;
            padding: 15px;
            border-radius: 5px;
            margin-top: 10px;
            font-size: 13px;
            color: #856404;
        }
        
        .modal {
            display: none;
            position: fixed;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            background: rgba(0,0,0,0.5);
            z-index: 1000;
            align-items: center;
            justify-content: center;
        }
        
        .modal.active {
            display: flex;
        }
        
        .modal-content {
            background: white;
            padding: 30px;
            border-radius: 15px;
            max-width: 500px;
            width: 90%;
        }
        
        .modal h3 {
            margin-bottom: 20px;
            color: #333;
        }
        
        .close-modal {
            float: right;
            font-size: 24px;
            cursor: pointer;
            color: #999;
        }
        
        .alert {
            padding: 15px;
            border-radius: 8px;
            margin-bottom: 20px;
            display: none;
        }
        
        .alert.success {
            background: #d4edda;
            color: #155724;
            border-left: 4px solid #28a745;
        }
        
        .alert.error {
            background: #f8d7da;
            color: #721c24;
            border-left: 4px solid #dc3545;
        }
        
        .alert.active {
            display: block;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1> WhatsApp Message Sender</h1>
            <p>Kirim pesan WhatsApp dengan mudah menggunakan template</p>
        </div>
        
        <div class="content">
            <div id="alert" class="alert"></div>
            
            <div class="info-box">
                <p><strong> Cara Penggunaan:</strong></p>
                <p>1. Pilih template atau buat pesan custom</p>
                <p>2. Isi nomor WhatsApp (format: 628xxxxxxxxxx)</p>
                <p>3. Sesuaikan pesan Anda</p>
                <p>4. Klik "Kirim ke WhatsApp"</p>
            </div>
            
            <div class="form-group">
                <label>Pilih Template Pesan:</label>
                <div class="template-buttons" id="templateButtons"></div>
            </div>
            
            <form id="messageForm">
                <div class="form-group">
                    <label for="phoneNumber">Nomor WhatsApp:</label>
                    <input type="tel" id="phoneNumber" placeholder="628123456789" required>
                    <small style="color: #666; font-size: 12px;">Gunakan format: 628xxxxxxxxxx (tanpa +)</small>
                </div>
                
                <div class="form-group">
                    <label for="message">Pesan:</label>
                    <textarea id="message" placeholder="Tulis pesan Anda di sini..." required></textarea>
                    <div id="templateVars" class="template-variables" style="display: none;"></div>
                </div>
                
                <div class="form-group">
                    <label for="link">Link (Opsional):</label>
                    <input type="text" id="link" placeholder="https://contoh.com">
                </div>
                
                <div class="button-group">
                    <button type="button" class="btn-secondary" onclick="openTemplateModal()">+ Buat Template Baru</button>
                    <button type="submit" class="btn-primary"> Kirim ke WhatsApp</button>
                </div>
            </form>
        </div>
    </div>
    
    <div id="templateModal" class="modal">
        <div class="modal-content">
            <span class="close-modal" onclick="closeTemplateModal()">&times;</span>
            <h3>Buat Template Baru</h3>
            <div class="form-group">
                <label for="templateName">Nama Template:</label>
                <input type="text" id="templateName" placeholder="Nama template">
            </div>
            <div class="form-group">
                <label for="templateContent">Isi Template:</label>
                <textarea id="templateContent" placeholder="Gunakan {variabel} untuk placeholder"></textarea>
                <small style="color: #666; font-size: 12px;">Contoh: Halo {nama}, promo {diskon}%</small>
            </div>
            <div class="button-group">
                <button class="btn-secondary" onclick="closeTemplateModal()">Batal</button>
                <button class="btn-primary" onclick="saveTemplate()">Simpan Template</button>
            </div>
        </div>
    </div>
    
    <script>
        let templates = [];
        let selectedTemplateId = null;
        
        async function loadTemplates() {
            try {
                const response = await fetch('/api/templates');
                templates = await response.json();
                renderTemplateButtons();
            } catch (error) {
                console.error('Error loading templates:', error);
            }
        }
        
        function renderTemplateButtons() {
            const container = document.getElementById('templateButtons');
            container.innerHTML = '';
            
            templates.forEach(template => {
                const btn = document.createElement('button');
                btn.type = 'button';
                btn.className = 'template-btn';
                btn.textContent = template.name;
                btn.onclick = () => selectTemplate(template.id);
                container.appendChild(btn);
            });
        }
        
        function selectTemplate(id) {
            selectedTemplateId = id;
            const template = templates.find(t => t.id === id);
            
            document.querySelectorAll('.template-btn').forEach(btn => {
                btn.classList.remove('active');
            });
            event.target.classList.add('active');
            
            document.getElementById('message').value = template.content;
            
            const vars = template.content.match(/\{([^}]+)\}/g);
            if (vars && template.content) {
                const varsDiv = document.getElementById('templateVars');
                varsDiv.style.display = 'block';
                varsDiv.innerHTML = '<strong>Variabel tersedia:</strong> ' + vars.join(', ') + 
                    '<br><small>Ganti variabel dengan teks yang sesuai sebelum mengirim</small>';
            } else {
                document.getElementById('templateVars').style.display = 'none';
            }
        }
        
        document.getElementById('messageForm').addEventListener('submit', async (e) => {
            e.preventDefault();
            
            const phoneNumber = document.getElementById('phoneNumber').value.trim();
            let message = document.getElementById('message').value.trim();
            const link = document.getElementById('link').value.trim();
            
            if (!phoneNumber.match(/^62\d{9,13}$/)) {
                showAlert('Nomor WhatsApp harus dalam format 628xxxxxxxxxx', 'error');
                return;
            }
            
            if (link) {
                message += '\n\n' + link;
            }
            
            const encodedMessage = encodeURIComponent(message);
            const whatsappUrl = 'https://wa.me/' + phoneNumber + '?text=' + encodedMessage;
            
            try {
                await fetch('/api/send', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({
                        phone_number: phoneNumber,
                        message: message
                    })
                });
            } catch (error) {
                console.error('Error:', error);
            }
            
            window.open(whatsappUrl, '_blank');
            showAlert('Membuka WhatsApp... Silakan lanjutkan di aplikasi WhatsApp', 'success');
        });
        
        function openTemplateModal() {
            document.getElementById('templateModal').classList.add('active');
        }
        
        function closeTemplateModal() {
            document.getElementById('templateModal').classList.remove('active');
            document.getElementById('templateName').value = '';
            document.getElementById('templateContent').value = '';
        }
        
        async function saveTemplate() {
            const name = document.getElementById('templateName').value.trim();
            const content = document.getElementById('templateContent').value.trim();
            
            if (!name || !content) {
                showAlert('Nama dan isi template harus diisi', 'error');
                return;
            }
            
            try {
                const response = await fetch('/api/templates/save', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({
                        name: name,
                        content: content
                    })
                });
                
                if (response.ok) {
                    showAlert('Template berhasil disimpan!', 'success');
                    closeTemplateModal();
                    loadTemplates();
                }
            } catch (error) {
                showAlert('Gagal menyimpan template', 'error');
            }
        }
        
        function showAlert(message, type) {
            const alert = document.getElementById('alert');
            alert.textContent = message;
            alert.className = 'alert ' + type + ' active';
            
            setTimeout(() => {
                alert.classList.remove('active');
            }, 5000);
        }
        
        loadTemplates();
    </script>
</body>
</html>`

	fmt.Fprint(w, htmlContent)
}

func templatesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(templates)
}

func saveTemplateHandler(w http.ResponseWriter, r *http.Request) {
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

	template := MessageTemplate{
		ID:      fmt.Sprintf("%d", len(templates)+1),
		Name:    newTemplate.Name,
		Content: newTemplate.Content,
	}

	templates = append(templates, template)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(template)
}

func sendMessageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var form MessageForm
	if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Log pesan yang dikirim
	log.Printf("Mengirim pesan ke %s: %s", form.PhoneNumber, form.Message)

	// Generate WhatsApp URL
	encodedMessage := url.QueryEscape(form.Message)
	whatsappURL := fmt.Sprintf("https://wa.me/%s?text=%s",
		strings.TrimPrefix(form.PhoneNumber, "+"),
		encodedMessage)

	response := map[string]string{
		"status":       "success",
		"whatsapp_url": whatsappURL,
		"message":      "URL WhatsApp berhasil dibuat",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
