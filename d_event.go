package gore

const MAXEVENTS = 64

var events [64]event_t
var eventhead int32
var eventtail int32

// C documentation
//
//	//
//	// D_PostEvent
//	// Called by the I/O functions when input is detected
//	//
func d_PostEvent(ev *event_t) {
	events[eventhead] = *ev
	eventhead = (eventhead + 1) % MAXEVENTS
}

// Read an event from the queue.

func d_PopEvent() *event_t {
	// No more events waiting.
	if eventtail == eventhead {
		return nil
	}
	result := &events[eventtail]
	// Advance to the next event in the queue.
	eventtail = (eventtail + 1) % MAXEVENTS
	return result
}
