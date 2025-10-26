package pipeline

import "fmt"

type Step interface {
	Run() error
	Name() string
}

type Pipeline struct {
	Steps []Step
}

func NewPipeline(steps ...Step) *Pipeline {
	return &Pipeline{Steps: steps}
}

func (p *Pipeline) Run() error {
	for _, step := range p.Steps {
		fmt.Printf("=== RUNNING STEP: %s ===\n", step.Name())
		if err := step.Run(); err != nil {
			return fmt.Errorf("pipeline step %q failed: %w", step.Name(), err)
		}
		fmt.Printf("=== COMPLETED STEP: %s ===\n\n", step.Name())
	}
	return nil
}
