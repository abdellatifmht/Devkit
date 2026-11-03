package answers_test

import (
	"testing"

	"github.com/abdellatifmht/Brik/internal/answers"
	"github.com/abdellatifmht/Brik/internal/manifest"
)

func testManifest() *manifest.Manifest {
	return &manifest.Manifest{
		Questions: []manifest.Question{
			{
				ID:      "mode",
				Prompt:  "Mode",
				Type:    "select",
				Default: "simple",
				Options: []manifest.Option{
					{Value: "simple"},
					{Value: "advanced"},
				},
			},
			{
				ID:      "advanced_option",
				Prompt:  "Advanced option",
				Type:    "input",
				Default: "default",
				When:    `{{ eq .mode "advanced" }}`,
			},
			{
				ID:      "confirm",
				Prompt:  "Confirm",
				Type:    "confirm",
				Default: true,
			},
		},
	}
}

func TestResolveWithFormCollectsAnswers(t *testing.T) {
	var got []string
	form := func(questions []answers.FormQuestion, initial map[string]any) (map[string]any, error) {
		for _, q := range questions {
			got = append(got, q.Question.ID)
		}
		if initial["mode"] != "simple" {
			t.Fatalf("initial mode = %v", initial["mode"])
		}
		return map[string]any{
			"mode":            "advanced",
			"advanced_option": "chosen",
			"confirm":         false,
		}, nil
	}

	data, err := answers.ResolveWithForm(testManifest(), nil, nil, false, nil, form)
	if err != nil {
		t.Fatal(err)
	}
	if data["mode"] != "advanced" || data["advanced_option"] != "chosen" || data["confirm"] != false {
		t.Fatalf("unexpected data: %#v", data)
	}
	if len(got) != 3 {
		t.Fatalf("questions = %v", got)
	}
}

func TestResolveWithFormClearsInactiveAnswers(t *testing.T) {
	form := func(_ []answers.FormQuestion, _ map[string]any) (map[string]any, error) {
		// Simulate a user who answered advanced_option and then went back
		// to choose "simple".
		return map[string]any{
			"mode":            "simple",
			"advanced_option": "stale",
			"confirm":         true,
		}, nil
	}

	data, err := answers.ResolveWithForm(testManifest(), nil, nil, false, nil, form)
	if err != nil {
		t.Fatal(err)
	}
	if data["advanced_option"] != "" {
		t.Fatalf("inactive answer should be cleared, got %#v", data["advanced_option"])
	}
	if data["confirm"] != true {
		t.Fatalf("confirm = %v", data["confirm"])
	}
}

func TestResolveWithFormSkipsPresetQuestions(t *testing.T) {
	var got []string
	form := func(questions []answers.FormQuestion, initial map[string]any) (map[string]any, error) {
		for _, q := range questions {
			got = append(got, q.Question.ID)
		}
		if initial["mode"] != "simple" {
			t.Fatalf("preset mode = %v", initial["mode"])
		}
		return map[string]any{
			"advanced_option": "ignored",
			"confirm":         true,
		}, nil
	}

	data, err := answers.ResolveWithForm(
		testManifest(),
		nil,
		map[string]string{"mode": "simple"},
		false,
		nil,
		form,
	)
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range got {
		if id == "mode" {
			t.Fatal("preset question should not be asked")
		}
	}
	if data["mode"] != "simple" {
		t.Fatalf("mode = %v", data["mode"])
	}
	if data["advanced_option"] != "" {
		t.Fatalf("inactive answer should be cleared, got %#v", data["advanced_option"])
	}
}

func TestResolveWithFormIgnoresFormWhenYes(t *testing.T) {
	called := false
	form := func(_ []answers.FormQuestion, _ map[string]any) (map[string]any, error) {
		called = true
		return nil, nil
	}

	data, err := answers.ResolveWithForm(testManifest(), nil, nil, true, nil, form)
	if err != nil {
		t.Fatal(err)
	}
	if called {
		t.Fatal("form should not run with --yes")
	}
	if data["mode"] != "simple" || data["advanced_option"] != "" || data["confirm"] != true {
		t.Fatalf("unexpected defaults: %#v", data)
	}
}
