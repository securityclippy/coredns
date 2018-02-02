package up

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestUp(t *testing.T) {
	// We check a target 5 times, then check if that really happened. After which fails should have been reset to 0.
	const max = 5
	wg := sync.WaitGroup{}
	fails := int32(3)
	hits := int32(0)
	web := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&hits, 1)
		}))
	defer web.Close()

	pr := New()
	defer pr.Stop()
	i := 0
	wg.Add(1)
	upfunc := func(s string) bool {
		http.Get(s)
		i++
		if i >= max {
			atomic.StoreInt32(&fails, 0)
			wg.Done()
			return true
		}
		return false
	}

	pr.Start(web.URL, 5*time.Millisecond)
	pr.Do(upfunc) // Kicks off tests

	pr.Do(upfunc) // noop
	pr.Do(upfunc) // noop
	pr.Do(upfunc) // noop

	wg.Wait()

	if fails != 0 {
		t.Errorf("Expecting fails to be 0, got %d", fails)
	}
	if hits != max {
		t.Errorf("Expecting hits to be %d, got %d", max, hits)
	}

	// Reset values and run once-more
	i = 0
	fails = int32(3)
	hits = int32(0)
	wg.Add(1)

	pr.Do(upfunc)

	wg.Wait()

	if fails != 0 {
		t.Errorf("Expecting fails to be 0, got %d", fails)
	}
	if hits != max {
		t.Errorf("Expecting hits to be %d, got %d", max, hits)
	}
}
