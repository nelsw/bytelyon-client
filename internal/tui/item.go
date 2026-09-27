package tui

import "fmt"

// targetItem adapts a Target to the bubbles/list.Item and DefaultItem
// interfaces.
type targetItem struct {
	target Target
}

func (i targetItem) Title() string {
	if i.target.NeedsApp {
		return fmt.Sprintf("%s *", i.target.Name)
	}
	return i.target.Name
}

func (i targetItem) Description() string {
	desc := i.target.Description
	if i.target.Category != "" {
		return categoryStyle.Render("["+i.target.Category+"] ") + desc
	}
	return desc
}

func (i targetItem) FilterValue() string {
	return i.target.Name + " " + i.target.Description + " " + i.target.Category
}
