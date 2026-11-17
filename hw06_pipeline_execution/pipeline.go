package hw06pipelineexecution

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) (out Out)

func ExecutePipeline(in In, done In, stages ...Stage) Out {
	currentIn := in
	for _, stage := range stages {
		currentIn = stage(link(currentIn, done))
	}
	return link(currentIn, done)
}

func link(source In, done In) In {
	if done == nil {
		return source
	}
	destination := make(Bi)
	go func() {
		var destClosed bool
		closeDest := func() {
			if !destClosed {
				close(destination)
				destClosed = true
			}
		}
		defer func() {
			closeDest()
		}()
		for {
			select {
			case <-done:
				closeDest()
				drain(source)
				return
			case val, ok := <-source:
				if !ok {
					return
				}
				select {
				case <-done:
					closeDest()
					drain(source)
					return
				case destination <- val:
				}
			}
		}
	}()
	return destination
}

// drain reads source until it is closed, allowing upstream stages to finish.
func drain(source In) {
	for {
		if _, ok := <-source; !ok {
			return
		}
	}
}
