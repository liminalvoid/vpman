package wizard

type Wizard struct {
	awaitingCallback State
	sentAnswer       State

	currentState     State
}

func newWizard()
