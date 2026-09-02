package progress

import (
	"fmt"
	"sync"
	"testing"
)

func TestTrackerSnapshotCountsSourceStates(t *testing.T) {
	tracker := New()
	tracker.Expect("first", 1)
	tracker.Expect("second", 2)
	tracker.Expect("third", 3)

	tracker.Start("first", 1)
	tracker.Done("first")
	tracker.Start("second", 2)

	snap := tracker.Snapshot()
	if snap.Total != 3 || snap.Done != 1 || snap.Running != 1 {
		t.Fatalf("snapshot = %+v", snap)
	}
	if snap.Finished {
		t.Fatal("snapshot should not be finished")
	}
	if snap.Percent != 33.3 {
		t.Fatalf("percent = %v", snap.Percent)
	}
	if snap.Sources[0].Name != "first" || snap.Sources[0].Status != StatusDone {
		t.Fatalf("first source = %+v", snap.Sources[0])
	}
	if snap.Sources[2].Status != StatusPending {
		t.Fatalf("third source = %+v", snap.Sources[2])
	}
}

func TestTrackerReportsLoopProgress(t *testing.T) {
	tracker := New()
	tracker.Expect("looped", 7)
	tracker.Start("looped", 7)
	tracker.SetLoop("looped", 10)
	tracker.LoopStep("looped")
	tracker.LoopStep("looped")

	entry := tracker.Snapshot().Sources[0]
	if !entry.Loop || entry.LoopDone != 2 || entry.LoopTotal != 10 {
		t.Fatalf("entry = %+v", entry)
	}
	if entry.Percent != 20 {
		t.Fatalf("entry percent = %v", entry.Percent)
	}

	tracker.Done("looped")
	entry = tracker.Snapshot().Sources[0]
	if entry.LoopDone != 10 || entry.Percent != 100 {
		t.Fatalf("completed entry = %+v", entry)
	}
}

func TestTrackerFailAndFinish(t *testing.T) {
	tracker := New()
	tracker.Start("broken", 4)
	tracker.Fail("broken", 4, "boom")
	tracker.Finish()

	snap := tracker.Snapshot()
	// Une source en erreur est terminée : elle compte dans Done, comme
	// l'anneau la compte déjà dans Percent. Sans quoi le compteur reste sous
	// son total alors que le chargement est fini.
	if snap.Errors != 1 || snap.Done != 1 || snap.Done != snap.Total || !snap.Finished || snap.Percent != 100 {
		t.Fatalf("snapshot = %+v", snap)
	}
	if snap.Sources[0].Error != "boom" || snap.Sources[0].Status != StatusError {
		t.Fatalf("entry = %+v", snap.Sources[0])
	}

	// Une source déjà en erreur ne doit pas repasser en "done".
	tracker.Done("broken")
	if status := tracker.Snapshot().Sources[0].Status; status != StatusError {
		t.Fatalf("status = %s", status)
	}
}

func TestTrackerRegistersUnexpectedSourceAndKeepsVersion(t *testing.T) {
	tracker := New()
	version := tracker.Version()

	tracker.Start("discovered", 9)
	tracker.Done("discovered")

	if tracker.Version() == version {
		t.Fatal("version did not change")
	}
	snap := tracker.Snapshot()
	if snap.Total != 1 || snap.Done != 1 || snap.Sources[0].ID != 9 {
		t.Fatalf("snapshot = %+v", snap)
	}
}

func TestNilTrackerIsSafe(t *testing.T) {
	var tracker *Tracker

	tracker.Expect("a", 1)
	tracker.Start("a", 1)
	tracker.SetLoop("a", 2)
	tracker.LoopStep("a")
	tracker.Done("a")
	tracker.Fail("a", 1, "err")
	tracker.Finish()

	if snap := tracker.Snapshot(); !snap.Finished || len(snap.Sources) != 0 {
		t.Fatalf("snapshot = %+v", snap)
	}
	if tracker.Version() != 0 {
		t.Fatal("version should be zero")
	}
}

func TestTrackerConcurrentUpdates(t *testing.T) {
	tracker := New()
	tracker.Expect("looped", 1)
	tracker.Start("looped", 1)
	tracker.SetLoop("looped", 50)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tracker.LoopStep("looped")
			tracker.Snapshot()
		}()
	}
	wg.Wait()

	if done := tracker.Snapshot().Sources[0].LoopDone; done != 50 {
		t.Fatalf("loop done = %d", done)
	}
}

func TestTrackerLoopFailMarksSourcePartial(t *testing.T) {
	tracker := New()
	tracker.Expect("looped", 3)
	tracker.Start("looped", 3)
	tracker.SetLoop("looped", 3)
	tracker.LoopStep("looped")
	tracker.LoopFail("looped", "1", "first boom")
	tracker.LoopStep("looped")
	tracker.LoopFail("looped", "2", "second boom")
	tracker.LoopStep("looped")
	tracker.Done("looped")
	tracker.Finish()

	snap := tracker.Snapshot()
	// Les données sont disponibles (Done) et une erreur a eu lieu (Errors).
	if snap.Done != 1 || snap.Errors != 1 || snap.Percent != 100 {
		t.Fatalf("snapshot = %+v", snap)
	}
	entry := snap.Sources[0]
	if entry.Status != StatusPartial || entry.LoopErrors != 2 || entry.LoopDone != 3 {
		t.Fatalf("entry = %+v", entry)
	}
	// Le premier message est l'erreur de la source, le détail liste tout.
	if entry.Error != "first boom" {
		t.Fatalf("entry error = %q", entry.Error)
	}
	if len(entry.Failures) != 2 || entry.Failures[0] != (LoopFailure{Key: "1", Message: "first boom"}) || entry.Failures[1].Key != "2" {
		t.Fatalf("entry failures = %+v", entry.Failures)
	}

	// Une source partielle ne doit pas repasser en "done".
	tracker.Done("looped")
	if status := tracker.Snapshot().Sources[0].Status; status != StatusPartial {
		t.Fatalf("status = %s", status)
	}
}

func TestTrackerDoneWithoutLoopErrorsStaysDone(t *testing.T) {
	tracker := New()
	tracker.Start("looped", 1)
	tracker.SetLoop("looped", 2)
	tracker.LoopStep("looped")
	tracker.LoopStep("looped")
	tracker.Done("looped")

	snap := tracker.Snapshot()
	if snap.Errors != 0 || snap.Sources[0].Status != StatusDone || snap.Sources[0].LoopErrors != 0 {
		t.Fatalf("snapshot = %+v", snap)
	}
}

func TestTrackerLoopFailDetailIsBounded(t *testing.T) {
	tracker := New()
	tracker.Start("looped", 1)
	for i := 0; i < maxLoopFailures+10; i++ {
		tracker.LoopFail("looped", fmt.Sprint(i), "boom")
	}
	entry := tracker.Snapshot().Sources[0]
	if entry.LoopErrors != maxLoopFailures+10 || len(entry.Failures) != maxLoopFailures {
		t.Fatalf("entry = %d errors, %d failures", entry.LoopErrors, len(entry.Failures))
	}
}
