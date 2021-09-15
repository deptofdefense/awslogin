package op

import "time"

type Item struct {
	Uuid         string    `json:"uuid,omitempty"`
	TemplateUuid string    `json:"templateUuid,omitempty"`
	Trashed      string    `json:"trashed,omitempty"`
	CreatedAt    time.Time `json:"createdAt,omitempty"`
	UpdatedAt    time.Time `json:"updatedAt,omitempty"`
	ChangerUuid  string    `json:"changeUuid,omitempty"`
	ItemVersion  int       `json:"itemVersion,omitempty"`
	VaultUuid    string    `json:"vaultUuid,omitempty"`
	Details      Details   `json:"details,omitempty"`
	Overview     Overview  `json:"overview,omitempty"`
}

type Details struct {
	Fields          []Field       `json:"fields,omitempty"`
	NotesPlain      string        `json:"notesPlain,omitempty"`
	PasswordHistory []interface{} `json:"passwordHistory,omitempty"`
	Sections        []Section     `json:",omitempty"`
}

type Field struct {
	Designation string `json:"designation,omitempty"`
	Name        string `json:"name,omitempty"`
	Type        string `json:"type,omitempty"`
	Value       string `json:"value,omitempty"`
}

type Section struct {
	Name   string         `json:"name,omitempty"`
	Title  string         `json:"title,omitempty"`
	Fields []SectionField `json:"fields,omitempty"`
}

type SectionField struct {
	K string `json:"k,omitempty"`
	N string `json:"n,omitempty"`
	T string `json:"t,omitempty"`
	V string `json:"v,omitempty"`
}

type Overview struct {
	URLs  []URL    `json:"URLs,omitempty"`
	Ainfo string   `json:"ainfo,omitempty"`
	Ps    int      `json:"ps,omitempty"`
	Pbe   float64  `json:"pbe,omitempty"`
	Pgrng bool     `json:"pgrng,omitempty"`
	Tags  []string `json:"tags,omitempty"`
	Title string   `json:"title,omitempty"`
	Url   string   `json:"url,omitempty"`
}

type URL struct {
	L string `json:"l,omitempty"`
	U string `json:"u,omitempty"`
}
