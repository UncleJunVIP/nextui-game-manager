package shared

type TableColumns struct {
	FilenameHeader string `yaml:"-" json:"-"`
	FileSizeHeader string `yaml:"-" json:"-"`
	DateHeader     string `yaml:"-" json:"-"`
}
