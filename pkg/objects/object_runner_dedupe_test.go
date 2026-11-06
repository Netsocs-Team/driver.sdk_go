package objects

import (
	"sync"
	"testing"
)

type stubObject struct {
	RegistrableObject
	meta ObjectMetadata
	tag  string
}

func (s *stubObject) GetMetadata() ObjectMetadata { return s.meta }

func stub(domain, id, tag string) *stubObject {
	return &stubObject{meta: ObjectMetadata{Domain: domain, ObjectID: id}, tag: tag}
}

func domainObjects(t *testing.T, r *objectRunner, domain string) []RegistrableObject {
	t.Helper()
	raw, ok := r.objectsMap.Load(domain)
	if !ok {
		return nil
	}
	return raw.([]RegistrableObject)
}

// Re-provisioning a device (hub restart, reconnect) registers the same objects
// again; each action must still run on a single instance, the newest one.
func TestStoreObjectReplacesSameID(t *testing.T) {
	r := &objectRunner{}
	if err := r.storeObject(stub("cam", "cam.1", "old")); err != nil {
		t.Fatal(err)
	}
	if err := r.storeObject(stub("cam", "cam.2", "other")); err != nil {
		t.Fatal(err)
	}
	if err := r.storeObject(stub("cam", "cam.1", "new")); err != nil {
		t.Fatal(err)
	}
	objs := domainObjects(t, r, "cam")
	if len(objs) != 2 {
		t.Fatalf("got %d objects, want 2 (one per id)", len(objs))
	}
	for _, o := range objs {
		if o.GetMetadata().ObjectID == "cam.1" && o.(*stubObject).tag != "new" {
			t.Fatal("re-registration did not replace the old instance")
		}
	}
}

func TestStoreObjectKeepsDomainsApart(t *testing.T) {
	r := &objectRunner{}
	r.storeObject(stub("cam", "x.1", "a"))
	r.storeObject(stub("switch", "x.1", "b"))
	if len(domainObjects(t, r, "cam")) != 1 || len(domainObjects(t, r, "switch")) != 1 {
		t.Fatal("objects with the same id in different domains must both stay")
	}
}

// The old Load/append/Store lost registrations when two ran at once.
func TestStoreObjectConcurrentRegistrations(t *testing.T) {
	r := &objectRunner{}
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			r.storeObject(stub("cam", string(rune('A'+i)), "t"))
		}(i)
	}
	wg.Wait()
	if n := len(domainObjects(t, r, "cam")); n != 50 {
		t.Fatalf("lost registrations under concurrency: %d of 50", n)
	}
}
