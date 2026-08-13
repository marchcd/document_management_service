package pkg

import (
	"fmt"
	"strings"

	"github.com/yalosev/petrovich"
)

func InflectDative(fullName string) string {
	parts := strings.Fields(strings.TrimSpace(fullName))

	if len(parts) < 2 {
		return fullName
	}

	person := petrovich.Person{
		LastName:  parts[0],
		FirstName: parts[1],
	}

	if len(parts) >= 3 {
		person.MiddleName = parts[2]
	}

	person.Gender = detectGender(person.MiddleName, person.FirstName)

	result := petrovich.Transform(person, petrovich.Dative)

	out := fmt.Sprintf("%s %s", result.LastName, result.FirstName)
	if result.MiddleName != "" {
		out += " " + result.MiddleName
	}

	return out
}

func detectGender(patronymic, firstName string) petrovich.Gender {
	p := strings.ToLower(patronymic)
	switch {
	case strings.HasSuffix(p, "ович"),
		strings.HasSuffix(p, "евич"),
		strings.HasSuffix(p, "ич"):
		return petrovich.Male
	case strings.HasSuffix(p, "овна"),
		strings.HasSuffix(p, "евна"),
		strings.HasSuffix(p, "ична"),
		strings.HasSuffix(p, "инична"):
		return petrovich.Female
	}

	fn := strings.ToLower(firstName)
	if strings.HasSuffix(fn, "а") || strings.HasSuffix(fn, "я") {
		return petrovich.Female
	}
	return petrovich.Male
}
