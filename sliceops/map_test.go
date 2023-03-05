package sliceops

import (
	"strconv"
	"testing"
	"time"
)

func TestMap(t *testing.T) {
	items := []int{0, 1, 2, 3, 4, 5, 6}

	for i, item := range Map(items, strconv.Itoa) {
		if item != strconv.Itoa(i) {
			t.Fatalf("%s is not %d", item, items[i])
		}
	}
}

func TestMapAsync(t *testing.T) {
	items := []int{0, 1, 2, 3, 4, 5, 6}

	slowItoa := func(i int) string {
		time.Sleep(1 * time.Second)
		return strconv.Itoa((i))
	}

	before := time.Now()
	result := MapAsync(items, slowItoa)

	if time.Since(before) > 2*time.Second {
		t.Fatal("Took too long we should do this parallel here!")
	}

	for i, item := range result {
		if item != strconv.Itoa(i) {
			t.Fatalf("%s is not %d", item, items[i])
		}
	}
}
