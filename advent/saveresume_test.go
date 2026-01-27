package advent

import (
	"os"
	"path/filepath"
	"testing"
)

// TestSaveAndLoad tests saving and loading a game
func TestSaveAndLoad(t *testing.T) {
	// Create a game with some state
	game := NewGame(12345, "", "", "", false, false, false, nil)
	game.Loc = 5
	game.Turns = 42
	game.Holdng = 3

	// Save to temp file
	tmpDir := t.TempDir()
	saveFile := filepath.Join(tmpDir, "test.sav")

	if err := game.SaveToFile(saveFile); err != nil {
		t.Fatalf("SaveToFile failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(saveFile); os.IsNotExist(err) {
		t.Fatal("Save file was not created")
	}

	// Load into new game
	game2 := NewGame(0, "", "", "", false, false, false, nil)
	if err := game2.LoadFromFile(saveFile); err != nil {
		t.Fatalf("LoadFromFile failed: %v", err)
	}

	// Verify state was preserved
	if game2.Loc != game.Loc {
		t.Errorf("Loc: got %d, want %d", game2.Loc, game.Loc)
	}
	if game2.Turns != game.Turns {
		t.Errorf("Turns: got %d, want %d", game2.Turns, game.Turns)
	}
	if game2.Holdng != game.Holdng {
		t.Errorf("Holdng: got %d, want %d", game2.Holdng, game.Holdng)
	}
}

// TestLoadInvalidMagic tests loading a file with wrong magic number
func TestLoadInvalidMagic(t *testing.T) {
	tmpDir := t.TempDir()
	saveFile := filepath.Join(tmpDir, "invalid.sav")

	// Write invalid JSON
	content := `{"magic": "wrong", "version": 1, "game": {}}`
	if err := os.WriteFile(saveFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	game := &Game{}
	err := game.LoadFromFile(saveFile)
	if err == nil {
		t.Error("LoadFromFile should fail with invalid magic number")
	}
}

// TestLoadInvalidJSON tests loading a file with invalid JSON
func TestLoadInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	saveFile := filepath.Join(tmpDir, "invalid.sav")

	// Write invalid JSON
	if err := os.WriteFile(saveFile, []byte("not json"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	game := &Game{}
	err := game.LoadFromFile(saveFile)
	if err == nil {
		t.Error("LoadFromFile should fail with invalid JSON")
	}
}

// TestLoadNonexistent tests loading a file that doesn't exist
func TestLoadNonexistent(t *testing.T) {
	game := &Game{}
	err := game.LoadFromFile("/nonexistent/path/game.sav")
	if err == nil {
		t.Error("LoadFromFile should fail for non-existent file")
	}
}

// TestAutoSaveRotation tests that autosave rotates files correctly
func TestAutoSaveRotation(t *testing.T) {
	tmpDir := t.TempDir()
	saveFile := filepath.Join(tmpDir, "autosave.sav")

	game := NewGame(12345, "", "", "", false, false, true, nil)
	game.Settings.Autosave = true
	game.Settings.AutoSaveFileName = saveFile

	// Do 5 autosaves
	for i := 0; i < 5; i++ {
		game.Turns = int32(i * 10) // Change state each time
		if err := game.AutoSave(); err != nil {
			t.Fatalf("AutoSave %d failed: %v", i, err)
		}
	}

	// Check that we have the base file plus .1, .2, .3
	expectedFiles := []string{
		saveFile,
		saveFile + ".1",
		saveFile + ".2",
		saveFile + ".3",
	}

	for _, f := range expectedFiles {
		if _, err := os.Stat(f); os.IsNotExist(err) {
			t.Errorf("Expected file %s to exist", f)
		}
	}

	// .4 should not exist (we only keep 4)
	if _, err := os.Stat(saveFile + ".4"); !os.IsNotExist(err) {
		t.Error("File .4 should not exist (max 4 saves)")
	}
}

// TestAutoSaveDisabled tests that autosave does nothing when disabled
func TestAutoSaveDisabled(t *testing.T) {
	tmpDir := t.TempDir()
	saveFile := filepath.Join(tmpDir, "autosave.sav")

	game := &Game{}
	game.Settings.Autosave = false
	game.Settings.AutoSaveFileName = saveFile

	if err := game.AutoSave(); err != nil {
		t.Fatalf("AutoSave failed: %v", err)
	}

	// File should not exist
	if _, err := os.Stat(saveFile); !os.IsNotExist(err) {
		t.Error("AutoSave should not create file when disabled")
	}
}
