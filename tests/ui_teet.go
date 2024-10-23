package main_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	main "github.com/Moukrea/goresetit"
)

func TestShowLogo(t *testing.T) {
	output := captureOutput(func() {
		main.ShowLogo()
	})

	expectedLines := []string{
		".d88b",
		"8P www",
		"8b  d8",
		"`Y88P'",
	}

	for _, line := range expectedLines {
		if !strings.Contains(output, line) {
			t.Errorf("Logo should contain line '%s'", line)
		}
	}
}

func TestCommitModel(t *testing.T) {
	testCases := []struct {
		name          string
		inputKeys     []string
		expectedDone  bool
		expectedValue string
	}{
		{
			name:          "Valid commit message",
			inputKeys:     []string{"t", "e", "s", "t", " ", "c", "o", "m", "m", "i", "t", "enter"},
			expectedDone:  true,
			expectedValue: "test commit",
		},
		{
			name:          "Empty commit message",
			inputKeys:     []string{"enter"},
			expectedDone:  false,
			expectedValue: "",
		},
		{
			name:          "Cancel with escape",
			inputKeys:     []string{"t", "e", "s", "t", "esc"},
			expectedDone:  false,
			expectedValue: "test",
		},
		{
			name:          "Cancel with ctrl+c",
			inputKeys:     []string{"t", "e", "s", "t", "ctrl+c"},
			expectedDone:  false,
			expectedValue: "test",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			model := main.InitialCommitModel()
			
			// Simulate key presses
			for _, key := range tc.inputKeys {
				var msg tea.Msg
				switch key {
				case "enter":
					msg = tea.KeyMsg{Type: tea.KeyEnter}
				case "esc":
					msg = tea.KeyMsg{Type: tea.KeyEsc}
				case "ctrl+c":
					msg = tea.KeyMsg{Type: tea.KeyCtrlC}
				default:
					msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{rune(key[0])}}
				}
				var cmd tea.Cmd
				model, cmd = model.Update(msg)
				if cmd == tea.Quit {
					break
				}
			}

			if model.Done != tc.expectedDone {
				t.Errorf("Expected done to be %v, got %v", tc.expectedDone, model.Done)
			}

			if tc.expectedValue != "" && model.TextInput.Value() != tc.expectedValue {
				t.Errorf("Expected value '%s', got '%s'", tc.expectedValue, model.TextInput.Value())
			}
		})
	}
}

func TestConfirmModel(t *testing.T) {
	testCases := []struct {
		name         string
		key         string
		expectedYes bool
	}{
		{"Confirm with y", "y", true},
		{"Confirm with Y", "Y", true},
		{"Deny with n", "n", false},
		{"Deny with N", "N", false},
		{"Cancel with q", "q", false},
		{"Cancel with Q", "Q", false},
		{"Cancel with escape", "esc", false},
		{"Cancel with ctrl+c", "ctrl+c", false},
		{"Invalid key", "x", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			model := main.ConfirmModel{
				Question: "Test question",
			}

			var msg tea.Msg
			switch tc.key {
			case "esc":
				msg = tea.KeyMsg{Type: tea.KeyEsc}
			case "ctrl+c":
				msg = tea.KeyMsg{Type: tea.KeyCtrlC}
			default:
				msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tc.key)}
			}

			updatedModel, cmd := model.Update(msg)
			finalModel := updatedModel.(main.ConfirmModel)

			if finalModel.Answer != tc.expectedYes {
				t.Errorf("Expected answer %v for key %s, got %v", tc.expectedYes, tc.key, finalModel.Answer)
			}

			if tc.expectedYes || tc.key == "n" || tc.key == "N" {
				if cmd != tea.Quit {
					t.Error("Expected Quit command for definitive answer")
				}
			}
		})
	}
}

func TestPromptConfirmation(t *testing.T) {
	testCases := []struct {
		name        string
		dryRun      bool
		mockInput   string
		expected    bool
		expectError bool
	}{
		{
			name:        "Confirm dry run",
			dryRun:      true,
			mockInput:   "y",
			expected:    true,
			expectError: false,
		},
		{
			name:        "Cancel dry run",
			dryRun:      true,
			mockInput:   "n",
			expected:    false,
			expectError: false,
		},
		{
			name:        "Confirm actual run",
			dryRun:      false,
			mockInput:   "y",
			expected:    true,
			expectError: false,
		},
		{
			name:        "Cancel actual run",
			dryRun:      false,
			mockInput:   "n",
			expected:    false,
			expectError: false,
		},
		{
            name:        "Error case",
            dryRun:      false,
            mockInput:   "",
            expected:    false,
            expectError: true,
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // Mock tea.Program
            oldNewProgram := main.NewTeaProgram
            defer func() { main.NewTeaProgram = oldNewProgram }()

            main.NewTeaProgram = func(m tea.Model) *tea.Program {
                return &mockTeaProgram{
                    msgs:    []tea.Msg{tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tc.mockInput)}},
                    result:  "",
                    hasErr:  tc.expectError,
                }
            }

            result, err := main.PromptConfirmation(tc.dryRun)

            if tc.expectError {
                if err == nil {
                    t.Error("Expected error but got none")
                }
                return
            }

            if err != nil {
                t.Errorf("Unexpected error: %v", err)
            }

            if result != tc.expected {
                t.Errorf("Expected result %v, got %v", tc.expected, result)
            }
        })
    }
}

func TestPromptCommitMessage(t *testing.T) {
    testCases := []struct {
        name          string
        mockInput     string
        expectedMsg   string
        expectError   bool
    }{
        {
            name:        "Valid commit message",
            mockInput:   "test commit",
            expectedMsg: "test commit",
            expectError: false,
        },
        {
            name:        "Cancel input",
            mockInput:   "",
            expectedMsg: "",
            expectError: false,
        },
        {
            name:        "Error case",
            mockInput:   "",
            expectedMsg: "",
            expectError: true,
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // Mock tea.Program
            oldNewProgram := main.NewTeaProgram
            defer func() { main.NewTeaProgram = oldNewProgram }()

            main.NewTeaProgram = func(m tea.Model) *tea.Program {
                return &mockTeaProgram{
                    msgs:    []tea.Msg{tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tc.mockInput)}},
                    result:  tc.expectedMsg,
                    hasErr:  tc.expectError,
                }
            }

            msg, err := main.PromptCommitMessage()

            if tc.expectError {
                if err == nil {
                    t.Error("Expected error but got none")
                }
                return
            }

            if err != nil {
                t.Errorf("Unexpected error: %v", err)
            }

            if msg != tc.expectedMsg {
                t.Errorf("Expected message '%s', got '%s'", tc.expectedMsg, msg)
            }
        })
    }
}