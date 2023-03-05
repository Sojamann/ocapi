package sliceops

import (
	"strconv"
	"testing"
	"time"
)

func TestCombine(t *testing.T) {
	items := []int{0, 1, 2, 3, 4, 5, 6}

	for i, item := range Combine(items, strconv.Itoa) {
		if item.a != items[i] {
			t.Fatalf("%d is not %d", item.a, items[i])
		}

		if item.b != strconv.Itoa(item.a) {
			t.Fatalf("%s is not the string for %d", item.b, item.a)
		}
	}
}

func TestCombineAsny(t *testing.T) {
	items := []int{0, 1, 2, 3, 4, 5, 6}

	slowItoa := func(i int) string {
		time.Sleep(1 * time.Second)
		return strconv.Itoa((i))
	}

	before := time.Now()
	result := CombineAsync(items, slowItoa)

	if time.Since(before) > 2*time.Second {
		t.Fatal("Took too long we should do this parallel here!")
	}

	for i, item := range result {
		if item.a != items[i] {
			t.Fatalf("%d is not %d", item.a, items[i])
		}

		if item.b != strconv.Itoa(item.a) {
			t.Fatalf("%s is not the string for %d", item.b, item.a)
		}
	}
}
