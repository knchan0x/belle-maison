package cache

import (
	"fmt"
	"testing"
)

var key, val = "test", "yes! test"

func TestAdd(t *testing.T) {
	Add(key, val, NEVER_EXPIRED)
}

func TestGet(t *testing.T) {
	TestAdd(t)
	got, ok := Get(key)
	if !ok {
		t.Errorf("cannot get value from key %s", key)
	}
	if got.(string) != val {
		t.Errorf("got %q, wanted %q", got, val)
	}
}

func TestDelete(t *testing.T) {
	TestAdd(t)
	TestGet(t)
	Delete(key)
	got, ok := Get(key)
	if got != nil || ok {
		t.Errorf("key %s has not deleted", key)
	}
}

func TestMultiRoutineAdd(t *testing.T) {
	for i := 0; i < 10; i++ {
		go func(i int) {
			Add(fmt.Sprintf("%s - %d", key, i), fmt.Sprintf("%s - %d", val, i), NEVER_EXPIRED)
		}(i)
	}
}

func TestMultiRoutineGet(t *testing.T) {
	TestMultiRoutineAdd(t)

	for i := 0; i < 10; i++ {
		go func(i int) {
			got, ok := Get(fmt.Sprintf("%s - %d", key, i))
			if !ok {
				t.Errorf("cannot get value from key %s", fmt.Sprintf("%s - %d", key, i))
			}
			if got.(string) != fmt.Sprintf("%s - %d", val, i) {
				t.Errorf("got %q, wanted %q", got, fmt.Sprintf("%s - %d", val, i))
			}
		}(i)
	}
}

func TestMultiRoutineDelete(t *testing.T) {
	TestMultiRoutineAdd(t)
	TestMultiRoutineGet(t)

	for i := 0; i < 10; i++ {
		go func(i int) {
			Delete(fmt.Sprintf("%s - %d", key, i))
			got, ok := Get(fmt.Sprintf("%s - %d", key, i))
			if got != nil || ok {
				t.Errorf("key %s has not deleted", fmt.Sprintf("%s - %d", key, i))
			}
		}(i)
	}
}
