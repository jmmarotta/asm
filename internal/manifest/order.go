package manifest

import "sort"

func SortedSkills(skills []Skill) []Skill {
	ordered := append([]Skill(nil), skills...)
	SortSkills(ordered)
	return ordered
}

func SortSkills(skills []Skill) {
	sort.Slice(skills, func(i, j int) bool {
		left := skills[i]
		right := skills[j]
		if left.Name != right.Name {
			return left.Name < right.Name
		}
		if left.Subdir != right.Subdir {
			return left.Subdir < right.Subdir
		}
		if left.Origin != right.Origin {
			return left.Origin < right.Origin
		}
		return left.Version < right.Version
	})
}
