package cache

import (
	"fmt"
	"sync"
	"testing"
)

var key, val = "test", "yes! test"

func Test_add(t *testing.T) {
	Add(key, val, 0)
}

func Test_get(t *testing.T) {
	Test_add(t)
	got, ok := Get(key)
	if !ok {
		t.Errorf("cannot get value from key %s", key)
	}
	if got.(string) != val {
		t.Errorf("got %q, wanted %q", got, val)
	}
}

func Test_delete(t *testing.T) {
	Test_add(t)
	Test_get(t)
	Delete(key)
	got, ok := Get(key)
	if got != nil || ok {
		t.Errorf("key %s has not deleted", key)
	}
}

func Test_add_MultiRoutine(t *testing.T) {
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			Add(fmt.Sprintf("%s - %d", key, i), fmt.Sprintf("%s - %d", val, i), 0)
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func Test_get_MultiRoutine(t *testing.T) {
	Test_add_MultiRoutine(t)

	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			got, ok := Get(fmt.Sprintf("%s - %d", key, i))
			if !ok {
				t.Errorf("cannot get value from key %s", fmt.Sprintf("%s - %d", key, i))
			}
			if got.(string) != fmt.Sprintf("%s - %d", val, i) {
				t.Errorf("got %q, wanted %q", got, fmt.Sprintf("%s - %d", val, i))
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func Test_delete_MultiRoutine(t *testing.T) {
	Test_add_MultiRoutine(t)
	Test_get_MultiRoutine(t)

	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			Delete(fmt.Sprintf("%s - %d", key, i))
			got, ok := Get(fmt.Sprintf("%s - %d", key, i))
			if got != nil || ok {
				t.Errorf("key %s has not deleted", fmt.Sprintf("%s - %d", key, i))
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func TestNew(t *testing.T) {
	c := New("InMemory")
	if c == nil {
		t.Error()
	}
}
