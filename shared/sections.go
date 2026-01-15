package shared

type Section struct {
	Name string `yaml:"section_name,omitempty" json:"section_name,omitempty"`

	SystemTag      string `yaml:"system_tag,omitempty" json:"system_tag,omitempty"`
	LocalDirectory string `yaml:"local_directory,omitempty" json:"local_directory,omitempty"`

	HostSubdirectory string `yaml:"host_subdirectory,omitempty" json:"host_subdirectory,omitempty"`

	RomMPlatformID string `yaml:"romm_platform_id,omitempty" json:"romm_platform_id,omitempty"`

	CollectionFilePath string `yaml:"collection_file_path,omitempty" json:"collection_file_path,omitempty"`
}
