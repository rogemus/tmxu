package utils

import "testing"

type testItem struct {
	ID string
}

func (i testItem) Eq(c Comperable) bool {
	o, ok := c.(testItem)
	return ok && i.ID == o.ID
}

type otherItem struct {
	ID string
}

func (i otherItem) Eq(c Comperable) bool {
	o, ok := c.(otherItem)
	return ok && i.ID == o.ID
}

func TestContain(t *testing.T) {
	tests := []struct {
		name  string
		slice []testItem
		el    testItem
		want  bool
	}{
		{"found", []testItem{{"test-item-1"}, {"test-item-2"}}, testItem{"test-item-2"}, true},
		{"found first", []testItem{{"test-item-1"}, {"test-item-2"}}, testItem{"test-item-1"}, true},
		{"not found", []testItem{{"test-item-1"}, {"test-item-2"}}, testItem{"test-item-3"}, false},
		{"empty slice", []testItem{}, testItem{"test-item-1"}, false},
		{"nil slice", nil, testItem{"test-item-1"}, false},
		{"single element found", []testItem{{"test-item-1"}}, testItem{"test-item-1"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Contain(tt.slice, tt.el)
			if got != tt.want {
				t.Errorf("Contain(%v, %v) = %v, want %v", tt.slice, tt.el, got, tt.want)
			}
		})
	}
}

func TestEqDifferentType(t *testing.T) {
	if (testItem{"test-item-1"}).Eq(otherItem{"test-item-1"}) {
		t.Error("Eq returned true for a different type with the same ID")
	}
}
