package dispatch

import "testing"

func TestDrainConsumesOnlyCurrentEpoch(t *testing.T) {
	buffer := NewBuffer()
	stationID := "station-1"

	if _, err := buffer.Enqueue(stationID, 1, 1, 20.0); err != nil {
		t.Fatalf("enqueue old epoch: %v", err)
	}
	if _, err := buffer.Enqueue(stationID, 2, 1, 21.0); err != nil {
		t.Fatalf("enqueue current epoch: %v", err)
	}
	if _, err := buffer.Enqueue(stationID, 2, 2, 22.0); err != nil {
		t.Fatalf("enqueue current epoch: %v", err)
	}
	if _, err := buffer.Enqueue(stationID, 3, 1, 23.0); err != nil {
		t.Fatalf("enqueue future epoch: %v", err)
	}

	ready, retired, pending := buffer.Drain(stationID, 2)

	if len(retired) != 1 || retired[0].Epoch != 1 {
		t.Fatalf("expected 1 retired command from epoch 1, got %+v", retired)
	}
	if len(ready) != 2 {
		t.Fatalf("expected 2 ready commands from epoch 2, got %+v", ready)
	}
	for _, command := range ready {
		if command.Epoch != 2 {
			t.Fatalf("ready command has epoch %d, want 2", command.Epoch)
		}
	}
	if ready[0].Sequence > ready[1].Sequence {
		t.Fatalf("ready commands not ordered by sequence: %+v", ready)
	}
	if len(pending) != 1 || pending[0].Epoch != 3 {
		t.Fatalf("expected 1 pending command from epoch 3, got %+v", pending)
	}

	// Future command must stay buffered, still awaiting confirmation.
	remaining := buffer.Pending(stationID)
	if len(remaining) != 1 || remaining[0].Epoch != 3 {
		t.Fatalf("expected future command to remain buffered, got %+v", remaining)
	}
}

func TestDrainReleasesFutureCommandsOnceEpochConfirmed(t *testing.T) {
	buffer := NewBuffer()
	stationID := "station-1"

	if _, err := buffer.Enqueue(stationID, 3, 1, 23.0); err != nil {
		t.Fatalf("enqueue future epoch: %v", err)
	}

	ready, _, pending := buffer.Drain(stationID, 2)
	if len(ready) != 0 || len(pending) != 1 {
		t.Fatalf("epoch 2 drain: ready=%+v pending=%+v", ready, pending)
	}

	ready, retired, pending := buffer.Drain(stationID, 3)
	if len(ready) != 1 || ready[0].Epoch != 3 {
		t.Fatalf("expected epoch 3 command ready after confirmation, got %+v", ready)
	}
	if len(retired) != 0 || len(pending) != 0 {
		t.Fatalf("unexpected retired=%+v pending=%+v", retired, pending)
	}
	if got := buffer.Pending(stationID); len(got) != 0 {
		t.Fatalf("expected buffer empty after confirmation, got %+v", got)
	}
}
