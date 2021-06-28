package src

func MakeUnboundedQueue() (chan<- Task, <-chan Task) {
	in := make(chan Task)
	out := make(chan Task)

	go func() {

		outCh := func() chan Task {
			if len(inQueue) == 0 {
				return nil
			}

			return out
		}

		cur := func() Task {
			if len(inQueue) == 0 {
				return Task{}
			}

			return inQueue[0]
		}

		for len(inQueue) > 0 || in != nil {
			select {
			case oc, ok := <-in:
				if !ok {
					in = nil
				} else {
					inQueue = append(inQueue, oc)
				}
			case outCh() <- cur():
				if out != nil {
					inQueue = inQueue[1:]
				}
			}
		}

		close(out)
	}()

	return in, out
}
