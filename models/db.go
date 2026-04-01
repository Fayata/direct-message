package models

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	_ "modernc.org/sqlite"
)

var db *sql.DB

func init() {
	var err error
	db, err = sql.Open("sqlite", "data.db")
	if err != nil {
		panic(err)
	}

	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		address TEXT NOT NULL,
		gender TEXT NOT NULL,
		birth_date TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL
	);`); err != nil {
		panic(err)
	}
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS form_settings (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL
	);`); err != nil {
		panic(err)
	}
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS form_fields (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		label TEXT NOT NULL,
		name TEXT NOT NULL,
		field_type TEXT NOT NULL,
		required INTEGER NOT NULL DEFAULT 1,
		options TEXT,
		sort_order INTEGER NOT NULL DEFAULT 0
	);`); err != nil {
		panic(err)
	}
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS form_submissions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		data TEXT NOT NULL,
		created_at TEXT NOT NULL
	);`); err != nil {
		panic(err)
	}
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS form2_settings (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL
	);`); err != nil {
		panic(err)
	}
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS form2_fields (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		label TEXT NOT NULL,
		name TEXT NOT NULL,
		field_type TEXT NOT NULL,
		required INTEGER NOT NULL DEFAULT 1,
		options TEXT,
		sort_order INTEGER NOT NULL DEFAULT 0
	);`); err != nil {
		panic(err)
	}
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS form2_submissions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		data TEXT NOT NULL,
		created_at TEXT NOT NULL
	);`); err != nil {
		panic(err)
	}
	// Default form settings
	for k, v := range map[string]string{
		"form_title":           "Form Data Diri",
		"form_subtitle":         "Silakan lengkapi data Anda.",
		"form_success_message":  "Terima kasih, data Anda berhasil disimpan. Anda dapat menutup halaman ini.",
		"form_help_text":        "Pastikan data yang Anda isi sudah benar sebelum menekan tombol submit.",
	} {
		_, _ = db.Exec(`INSERT OR IGNORE INTO form_settings (key, value) VALUES (?, ?)`, k, v)
	}
	// Default form2 settings
	for k, v := range map[string]string{
		"form_title":           "AwanGold - Order Emas",
		"form_subtitle":         "Isi data untuk melanjutkan pembelian emas realtime.",
		"form_success_message":  "Terima kasih, order AwanGold Anda berhasil disimpan. Tim kami akan menghubungi Anda.",
		"form_help_text":        "Pastikan data yang Anda isi sudah benar sebelum menekan tombol submit.",
	} {
		_, _ = db.Exec(`INSERT OR IGNORE INTO form2_settings (key, value) VALUES (?, ?)`, k, v)
	}
	// Seed default form2 fields if none exist
	var n2 int
	_ = db.QueryRow(`SELECT COUNT(*) FROM form2_fields`).Scan(&n2)
	if n2 == 0 {
		seed2 := []struct {
			label, name, fieldType, options string
			sortOrder                        int
		}{
			{"Nama Lengkap", "name", "text", "", 0},
			{"Nomor WhatsApp", "phone", "tel", "", 1},
			{"Email", "email", "email", "", 2},
			{"Kota Pengiriman", "city", "text", "", 3},
			{"Alamat Lengkap", "address", "textarea", "", 4},
			{"Berat Emas (gram)", "gold_weight", "number", "", 5},
			{"Harga per Gram", "gold_price", "number", "", 6},
			{"Total Harga", "total_price", "number", "", 7},
		}
		for _, s := range seed2 {
			_, _ = db.Exec(`INSERT INTO form2_fields (label, name, field_type, required, options, sort_order) VALUES (?, ?, ?, 1, ?, ?)`,
				s.label, s.name, s.fieldType, s.options, s.sortOrder)
		}
	}
	// Seed default form fields if none exist
	var n int
	_ = db.QueryRow(`SELECT COUNT(*) FROM form_fields`).Scan(&n)
	if n == 0 {
		seed := []struct {
			label, name, fieldType, options string
			sortOrder                        int
		}{
			{"Nama Lengkap", "name", "text", "", 0},
			{"Nomor KTP", "nik", "text", "", 1},
			{"Nomor Handphone", "phone", "text", "", 2},
			{"Email", "email", "text", "", 3},
			{"Jenis Kelamin", "gender", "select", "Laki-laki,Perempuan", 4},
			{"Tanggal Lahir", "birth_date", "date", "", 5},
			{"Alamat Lengkap", "address", "textarea", "", 6},
			{"Jumlah Pinjaman", "loan_amount", "text", "", 7},
			{"Tenor (bulan)", "tenor", "text", "", 8},
			{"Tujuan Pinjaman", "purpose", "text", "", 9},
			{"Pekerjaan", "job", "text", "", 10},
			{"Penghasilan Bulanan", "income", "text", "", 11},
		}
		for _, s := range seed {
			_, _ = db.Exec(`INSERT INTO form_fields (label, name, field_type, required, options, sort_order) VALUES (?, ?, ?, 1, ?, ?)`,
				s.label, s.name, s.fieldType, s.options, s.sortOrder)
		}
	}
	ensureLoanFormFields()

	// Migrate existing users to form_submissions once
	var subCount int
	_ = db.QueryRow(`SELECT COUNT(*) FROM form_submissions`).Scan(&subCount)
	if subCount == 0 {
		rows, err := db.Query(`SELECT id, name, address, gender, birth_date, created_at FROM users`)
		if err == nil {
			for rows.Next() {
				var id int64
				var name, address, gender, birthDate, createdAt string
				if err := rows.Scan(&id, &name, &address, &gender, &birthDate, &createdAt); err != nil {
					continue
				}
				data := map[string]string{"name": name, "address": address, "gender": gender, "birth_date": birthDate}
				js, _ := json.Marshal(data)
				_, _ = db.Exec(`INSERT INTO form_submissions (data, created_at) VALUES (?, ?)`, string(js), createdAt)
			}
			rows.Close()
		}
	}
}

func GetForm2Settings(ctx context.Context) (map[string]string, error) {
	rows, err := db.QueryContext(ctx, `SELECT key, value FROM form2_settings`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	m := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		m[k] = v
	}
	return m, rows.Err()
}

func SetForm2Setting(ctx context.Context, key, value string) error {
	_, err := db.ExecContext(ctx, `INSERT INTO form2_settings (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value`, key, value)
	return err
}

func ListForm2Fields(ctx context.Context) ([]FormField, error) {
	rows, err := db.QueryContext(ctx, `SELECT id, label, name, field_type, required, options, sort_order FROM form2_fields ORDER BY sort_order ASC, id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []FormField
	for rows.Next() {
		var f FormField
		var req int
		if err := rows.Scan(&f.ID, &f.Label, &f.Name, &f.FieldType, &req, &f.Options, &f.SortOrder); err != nil {
			return nil, err
		}
		f.Required = req != 0
		list = append(list, f)
	}
	return list, rows.Err()
}

func GetForm2FieldByID(ctx context.Context, id int64) (FormField, error) {
	var f FormField
	var req int
	err := db.QueryRowContext(ctx, `SELECT id, label, name, field_type, required, options, sort_order FROM form2_fields WHERE id = ?`, id).
		Scan(&f.ID, &f.Label, &f.Name, &f.FieldType, &req, &f.Options, &f.SortOrder)
	if err != nil {
		return f, err
	}
	f.Required = req != 0
	return f, nil
}

func CreateForm2Field(ctx context.Context, f *FormField) error {
	res, err := db.ExecContext(ctx, `INSERT INTO form2_fields (label, name, field_type, required, options, sort_order) VALUES (?, ?, ?, ?, ?, ?)`,
		f.Label, f.Name, f.FieldType, boolToInt(f.Required), f.Options, f.SortOrder)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	f.ID = id
	return nil
}

func UpdateForm2Field(ctx context.Context, f *FormField) error {
	_, err := db.ExecContext(ctx, `UPDATE form2_fields SET label=?, name=?, field_type=?, required=?, options=?, sort_order=? WHERE id=?`,
		f.Label, f.Name, f.FieldType, boolToInt(f.Required), f.Options, f.SortOrder, f.ID)
	return err
}

func DeleteForm2Field(ctx context.Context, id int64) error {
	_, err := db.ExecContext(ctx, `DELETE FROM form2_fields WHERE id = ?`, id)
	return err
}

func SaveForm2Submission(ctx context.Context, data map[string]string) (int64, error) {
	js, err := json.Marshal(data)
	if err != nil {
		return 0, err
	}
	res, err := db.ExecContext(ctx, `INSERT INTO form2_submissions (data, created_at) VALUES (?, ?)`, string(js), time.Now().Format(time.RFC3339))
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func ListForm2Submissions(ctx context.Context) ([]Submission, error) {
	rows, err := db.QueryContext(ctx, `SELECT id, data, created_at FROM form2_submissions ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Submission
	for rows.Next() {
		var s Submission
		var dataStr, createdAt string
		if err := rows.Scan(&s.ID, &dataStr, &createdAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(dataStr), &s.Data)
		if s.Data == nil {
			s.Data = make(map[string]string)
		}
		for _, layout := range []string{time.RFC3339, "2006-01-02 15:04:05"} {
			if t, err := time.Parse(layout, createdAt); err == nil {
				s.CreatedAt = t
				break
			}
		}
		list = append(list, s)
	}
	return list, rows.Err()
}

func CountForm2Submissions(ctx context.Context) (int, error) {
	var n int
	err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM form2_submissions`).Scan(&n)
	return n, err
}

func ListForm2SubmissionsPage(ctx context.Context, limit, offset int) ([]Submission, error) {
	rows, err := db.QueryContext(ctx, `SELECT id, data, created_at FROM form2_submissions ORDER BY created_at DESC LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Submission
	for rows.Next() {
		var s Submission
		var dataStr, createdAt string
		if err := rows.Scan(&s.ID, &dataStr, &createdAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(dataStr), &s.Data)
		if s.Data == nil {
			s.Data = make(map[string]string)
		}
		for _, layout := range []string{time.RFC3339, "2006-01-02 15:04:05"} {
			if t, err := time.Parse(layout, createdAt); err == nil {
				s.CreatedAt = t
				break
			}
		}
		list = append(list, s)
	}
	return list, rows.Err()
}

func DeleteForm2Submission(ctx context.Context, id int64) error {
	_, err := db.ExecContext(ctx, `DELETE FROM form2_submissions WHERE id = ?`, id)
	return err
}

// ensureLoanFormFields menambahkan field pinjaman ke DB lama yang hanya punya 4 field awal.
func ensureLoanFormFields() {
	extra := []struct {
		label, name, fieldType, options string
	}{
		{"Nomor KTP", "nik", "text", ""},
		{"Nomor Handphone", "phone", "text", ""},
		{"Email", "email", "text", ""},
		{"Jumlah Pinjaman", "loan_amount", "text", ""},
		{"Tenor (bulan)", "tenor", "text", ""},
		{"Tujuan Pinjaman", "purpose", "text", ""},
		{"Pekerjaan", "job", "text", ""},
		{"Penghasilan Bulanan", "income", "text", ""},
	}
	for _, f := range extra {
		var c int
		if err := db.QueryRow(`SELECT COUNT(*) FROM form_fields WHERE name = ?`, f.name).Scan(&c); err != nil {
			continue
		}
		if c > 0 {
			continue
		}
		var maxSort int
		_ = db.QueryRow(`SELECT COALESCE(MAX(sort_order), -1) FROM form_fields`).Scan(&maxSort)
		_, _ = db.Exec(`INSERT INTO form_fields (label, name, field_type, required, options, sort_order) VALUES (?, ?, ?, 1, ?, ?)`,
			f.label, f.name, f.fieldType, f.options, maxSort+1)
	}
}

type User struct {
	ID        int64
	Name      string
	Address   string
	Gender    string
	BirthDate string
	CreatedAt time.Time
}

func SaveUser(ctx context.Context, u *User) error {
	res, err := db.ExecContext(ctx, `
		INSERT INTO users (name, address, gender, birth_date, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, u.Name, u.Address, u.Gender, u.BirthDate, time.Now())
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err == nil {
		u.ID = id
	}
	u.CreatedAt = time.Now()
	return nil
}

func ListUsers(ctx context.Context) ([]User, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, name, address, gender, birth_date, created_at
		FROM users
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []User
	for rows.Next() {
		var u User
		var createdAt string
		if err := rows.Scan(&u.ID, &u.Name, &u.Address, &u.Gender, &u.BirthDate, &createdAt); err != nil {
			return nil, err
		}
		for _, layout := range []string{"2006-01-02 15:04:05.999999999-07:00", time.RFC3339, "2006-01-02 15:04:05"} {
			if t, err := time.Parse(layout, createdAt); err == nil {
				u.CreatedAt = t
				break
			}
		}
		list = append(list, u)
	}
	return list, rows.Err()
}

func DeleteUser(ctx context.Context, id int64) error {
	_, err := db.ExecContext(ctx, `DELETE FROM users WHERE id = ?`, id)
	return err
}

func GetFormSettings(ctx context.Context) (map[string]string, error) {
	rows, err := db.QueryContext(ctx, `SELECT key, value FROM form_settings`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	m := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		m[k] = v
	}
	return m, rows.Err()
}

func SetFormSetting(ctx context.Context, key, value string) error {
	_, err := db.ExecContext(ctx, `INSERT INTO form_settings (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value`, key, value)
	return err
}

// FormField: definisi field form yang dikelola admin
type FormField struct {
	ID        int64
	Label     string
	Name      string
	FieldType string
	Required  bool
	Options   string
	SortOrder int
}

func ListFormFields(ctx context.Context) ([]FormField, error) {
	rows, err := db.QueryContext(ctx, `SELECT id, label, name, field_type, required, options, sort_order FROM form_fields ORDER BY sort_order ASC, id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []FormField
	for rows.Next() {
		var f FormField
		var req int
		if err := rows.Scan(&f.ID, &f.Label, &f.Name, &f.FieldType, &req, &f.Options, &f.SortOrder); err != nil {
			return nil, err
		}
		f.Required = req != 0
		list = append(list, f)
	}
	return list, rows.Err()
}

func GetFormFieldByID(ctx context.Context, id int64) (FormField, error) {
	var f FormField
	var req int
	err := db.QueryRowContext(ctx, `SELECT id, label, name, field_type, required, options, sort_order FROM form_fields WHERE id = ?`, id).
		Scan(&f.ID, &f.Label, &f.Name, &f.FieldType, &req, &f.Options, &f.SortOrder)
	if err != nil {
		return f, err
	}
	f.Required = req != 0
	return f, nil
}

func CreateFormField(ctx context.Context, f *FormField) error {
	res, err := db.ExecContext(ctx, `INSERT INTO form_fields (label, name, field_type, required, options, sort_order) VALUES (?, ?, ?, ?, ?, ?)`,
		f.Label, f.Name, f.FieldType, boolToInt(f.Required), f.Options, f.SortOrder)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	f.ID = id
	return nil
}

func UpdateFormField(ctx context.Context, f *FormField) error {
	_, err := db.ExecContext(ctx, `UPDATE form_fields SET label=?, name=?, field_type=?, required=?, options=?, sort_order=? WHERE id=?`,
		f.Label, f.Name, f.FieldType, boolToInt(f.Required), f.Options, f.SortOrder, f.ID)
	return err
}

func DeleteFormField(ctx context.Context, id int64) error {
	_, err := db.ExecContext(ctx, `DELETE FROM form_fields WHERE id = ?`, id)
	return err
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// Submission: data yang diisi user (data = map field name -> value)
type Submission struct {
	ID        int64
	Data      map[string]string
	CreatedAt time.Time
}

func SaveSubmission(ctx context.Context, data map[string]string) (int64, error) {
	js, err := json.Marshal(data)
	if err != nil {
		return 0, err
	}
	res, err := db.ExecContext(ctx, `INSERT INTO form_submissions (data, created_at) VALUES (?, ?)`, string(js), time.Now().Format(time.RFC3339))
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func ListSubmissions(ctx context.Context) ([]Submission, error) {
	rows, err := db.QueryContext(ctx, `SELECT id, data, created_at FROM form_submissions ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Submission
	for rows.Next() {
		var s Submission
		var dataStr, createdAt string
		if err := rows.Scan(&s.ID, &dataStr, &createdAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(dataStr), &s.Data)
		if s.Data == nil {
			s.Data = make(map[string]string)
		}
		for _, layout := range []string{time.RFC3339, "2006-01-02 15:04:05"} {
			if t, err := time.Parse(layout, createdAt); err == nil {
				s.CreatedAt = t
				break
			}
		}
		list = append(list, s)
	}
	return list, rows.Err()
}

func CountSubmissions(ctx context.Context) (int, error) {
	var n int
	err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM form_submissions`).Scan(&n)
	return n, err
}

// CountSubmissionsSince menghitung submission form utama (Awan Peduli) sejak waktu tertentu.
func CountSubmissionsSince(ctx context.Context, since time.Time) (int, error) {
	cutoff := since.Format(time.RFC3339)
	var n int
	err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM form_submissions WHERE created_at >= ?`, cutoff).Scan(&n)
	return n, err
}

// CountForm2SubmissionsSince menghitung submission Awan Gold sejak waktu tertentu.
func CountForm2SubmissionsSince(ctx context.Context, since time.Time) (int, error) {
	cutoff := since.Format(time.RFC3339)
	var n int
	err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM form2_submissions WHERE created_at >= ?`, cutoff).Scan(&n)
	return n, err
}

// ListSubmissionsPage mengembalikan subset submissions (pagination sederhana).
// NOTE: filtering dilakukan di app layer (data JSON), jadi kita tetap baca data JSON.
func ListSubmissionsPage(ctx context.Context, limit, offset int) ([]Submission, error) {
	rows, err := db.QueryContext(ctx, `SELECT id, data, created_at FROM form_submissions ORDER BY created_at DESC LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Submission
	for rows.Next() {
		var s Submission
		var dataStr, createdAt string
		if err := rows.Scan(&s.ID, &dataStr, &createdAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(dataStr), &s.Data)
		if s.Data == nil {
			s.Data = make(map[string]string)
		}
		for _, layout := range []string{time.RFC3339, "2006-01-02 15:04:05"} {
			if t, err := time.Parse(layout, createdAt); err == nil {
				s.CreatedAt = t
				break
			}
		}
		list = append(list, s)
	}
	return list, rows.Err()
}

func DeleteSubmission(ctx context.Context, id int64) error {
	_, err := db.ExecContext(ctx, `DELETE FROM form_submissions WHERE id = ?`, id)
	return err
}

