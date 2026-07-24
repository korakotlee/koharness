package cmd

import (
	"bytes"
	"testing"
)

func TestRootCmdFlags(t *testing.T) {
	configFlag := RootCmd.PersistentFlags().Lookup("config")
	if configFlag == nil {
		t.Errorf("expected --config flag to be registered")
	}

	verboseFlag := RootCmd.PersistentFlags().Lookup("verbose")
	if verboseFlag == nil {
		t.Errorf("expected --verbose flag to be registered")
	}

	versionFlag := RootCmd.Flags().Lookup("version")
	if versionFlag == nil {
		t.Errorf("expected --version flag to be registered")
	}
}

func TestExecuteHelp(t *testing.T) {
	buf := new(bytes.Buffer)
	RootCmd.SetOut(buf)
	RootCmd.SetArgs([]string{"--help"})

	err := RootCmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error executing --help: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Errorf("expected non-empty help output")
	}
}
