package manifest

func (config Config) GitOriginVersions() map[string]string {
	origins := make(map[string]string)
	for _, skill := range config.Skills {
		if skill.Type != "git" {
			continue
		}
		origins[skill.Origin] = skill.Version
	}
	return origins
}
