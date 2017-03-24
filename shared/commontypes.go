package shared

import "fmt"

type CourseInfo struct {
	Name     string `json:"name"`
	Headline string `json:"headline"`
	Link     string `json:"link"`
	Art      string `json:"art"`
}

func (ci CourseInfo) String() string {
	return fmt.Sprintf("*%s*\n%s\n%s", ci.Name, ci.Headline, ci.Link)
}
