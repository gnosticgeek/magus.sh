package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"charm.land/huh/v2"
	"github.com/charmbracelet/x/ansi"
)

// Accessible mode is explicit: ordinary prompts can be read without a full-screen UI.
func accessibleSetup() bool { return os.Getenv("MAGUS_ACCESSIBLE") == "1" }

func runSetupForms(d Device, accessible bool, input io.Reader, output io.Writer) (Manifest, bool, error) {
	w := NewWizard(d)
	var accessibleInput *setupPromptReader
	if accessible {
		accessibleInput = &setupPromptReader{reader: bufio.NewReader(input)}
		input = accessibleInput
		output = setupPlainWriter{output}
	}

	run := func(groups ...*huh.Group) error {
		err := huh.NewForm(groups...).WithAccessible(accessible).WithInput(input).WithOutput(output).Run()
		if accessibleInput != nil && accessibleInput.err != nil {
			return accessibleInput.err
		}
		return err
	}
	choices := make([]string, len(w.Questions))
	bundles := append([]string{}, w.M.Choices.Bundles...)
	groups := make([]*huh.Group, 0, len(w.Questions)+1)
	descriptions := make([]string, len(w.Questions))
	for i, q := range w.Questions {
		options := make([]huh.Option[string], 0, len(q.Options))
		for _, o := range q.Options {
			options = append(options, huh.NewOption(o.Name+" — "+o.Note, o.ID))
			if w.Selected(q, o.ID) {
				choices[i] = o.ID
			}
		}
		if q.ID == "optimise" {
			built, pending := optimisationSummary(d)
			descriptions[i] = "Available:\n" + strings.Join(built, "\n")
			if len(pending) > 0 {
				descriptions[i] += "\nRecorded but not applied yet:\n" + strings.Join(pending, "\n")
			}
		}
		if q.ID == "theme" {
			descriptions[i] = "Recorded in your manifest; applying themes is not implemented yet."
		}
		var field huh.Field
		if q.Multi {
			field = huh.NewMultiSelect[string]().Title("Which bundles?").Description(descriptions[i]).Options(options...).Value(&bundles)
		} else {
			field = huh.NewSelect[string]().Title(q.Prompt).Description(descriptions[i]).Options(options...).Value(&choices[i])
		}
		groups = append(groups, huh.NewGroup(field))
	}
	applyChoices := func() {
		for i, q := range w.Questions {
			if !q.Multi {
				w.apply(q, choices[i])
			}
		}
		w.M.Choices.Bundles = append([]string{}, bundles...)
		sortBundles(w.M.Choices.Bundles)
	}
	review := func() string { applyChoices(); return setupReview(w) }
	confirmed := false
	confirm := huh.NewConfirm().Title("Write this manifest and apply the reviewed setup?").
		Affirmative("Apply setup").Negative("Cancel").Value(&confirmed)
	var err error
	if accessible {
		for i, group := range groups {
			if descriptions[i] != "" {
				fmt.Fprintln(output, descriptions[i])
			}
			if err = run(group); err != nil {
				return Manifest{}, false, err
			}
		}
		fmt.Fprintln(output, review())
		err = run(huh.NewGroup(confirm))
	} else {
		for {
			if err = run(groups...); err != nil {
				break
			}
			decision := "cancel"
			err = run(huh.NewGroup(huh.NewSelect[string]().Title("Review setup").Description(review()).Options(
				huh.NewOption("Apply reviewed setup", "apply"),
				huh.NewOption("Edit choices", "edit"),
				huh.NewOption("Cancel", "cancel")).Value(&decision)))
			if err != nil || decision != "edit" {
				confirmed = err == nil && decision == "apply"
				break
			}
		}
	}
	if errors.Is(err, huh.ErrUserAborted) {
		return Manifest{}, false, nil
	}
	if err != nil {
		return Manifest{}, false, fmt.Errorf("setup: %w", err)
	}
	applyChoices()

	return w.M, confirmed, nil
}

// Huh creates a scanner per prompt. Limit reads to one line so it cannot
// consume answers to subsequent prompts, and propagate input closure.
type setupPromptReader struct {
	reader *bufio.Reader
	err    error
}

func (r *setupPromptReader) Read(p []byte) (int, error) {
	for i := range p {
		b, err := r.reader.ReadByte()
		if err != nil {
			r.err = err
			return i, err
		}
		p[i] = b
		if b == '\n' {
			return i + 1, nil
		}
	}
	return len(p), nil
}

type setupPlainWriter struct{ io.Writer }

func (w setupPlainWriter) Write(p []byte) (int, error) {
	_, err := io.WriteString(w.Writer, ansi.Strip(string(p)))
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func setupReview(w *Wizard) string {
	return fmt.Sprintf("Terminal: %s\nBrowser: %s\nBundles: %s\nOptimisations: %t\nTheme: %s (recorded only)\n\nWrites ~/.config/magus/manifest.toml, then runs %d setup steps.\nRe-running repairs drift. Unimplemented choices remain recorded and are skipped.",
		w.M.Choices.Terminal, w.M.Choices.Browser, strings.Join(w.M.Choices.Bundles, ", "), w.M.Optimisations.Any(), w.M.Choices.Theme, len(StepsFor(w.M)))
}
