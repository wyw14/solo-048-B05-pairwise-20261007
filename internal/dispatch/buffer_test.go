package dispatch

import "testing"

func TestDrainConsumesOnlyCurrentEpoch(t *testing.T) {
	buffer := NewBuffer()
	for _, epoch := range []uint64{1, 2, 2, 3} {
		if _, err := buffer.Enqueue("st-1", epoch, epoch*10, 20.0); err != nil {
			t.Fatalf("enqueue epoch %d: %v", epoch, err)
		}
	}

	ready, retired := buffer.Drain("st-1", 2)
	if len(retired) != 1 || retired[0].Epoch != 1 {
		t.Fatalf("retired = %+v, want the single epoch-1 command", retired)
	}
	if len(ready) != 2 {
		t.Fatalf("ready = %+v, want the two epoch-2 commands", ready)
	}
	for _, command := range ready {
		if command.Epoch != 2 {
			t.Fatalf("ready contains future epoch %d", command.Epoch)
		}
	}

	pending := buffer.Pending("st-1")
	if len(pending) != 1 || pending[0].Epoch != 3 {
		t.Fatalf("pending = %+v, want the epoch-3 command still buffered", pending)
	}

	// Once epoch 3 is confirmed, the buffered future command becomes ready.
	ready, retired = buffer.Drain("st-1", 3)
	if len(retired) != 0 {
		t.Fatalf("retired = %+v, want none", retired)
	}
	if len(ready) != 1 || ready[0].Epoch != 3 {
		t.Fatalf("ready = %+v, want the epoch-3 command", ready)
	}
	if got := buffer.Pending("st-1"); len(got) != 0 {
		t.Fatalf("pending = %+v, want empty", got)
	}
}
