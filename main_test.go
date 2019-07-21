package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func readFromFile(fileName string) (string, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return "", err
	}
	defer f.Close()
	bytes, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

func TestCreateFileFromLabels(t *testing.T) {
	n := &nodeLabelsToFiles{config: &config{}}
	dir, err := ioutil.TempDir("", "nodeLabelsToFiles-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	n.config.directory = dir
	labels := map[string]string{
		"arch": "amd64",
		"failure-domain.beta.kubernetes.io/region": "us-east"}
	err = n.createFileFromLabels(labels)
	if err != nil {
		t.Error(err)
	}
	for fileName, expectedContent := range labels {
		actualContent, err := readFromFile(filepath.Join(n.config.directory,
			fileName))
		if err != nil {
			t.Error(err)
		}

		if expectedContent != actualContent {
			t.Errorf("Expected: %s, got: %s", expectedContent, actualContent)
		}
	}
}

func TestFilesToDelete(t *testing.T) {
	t.Run("Works When there is a file to delete", func(t *testing.T) {
		n := &nodeLabelsToFiles{config: &config{directory: "test/resources"}}
		labels := map[string]string{"bar": "", "baz": ""}
		expected := []string{"test/resources/foo"}
		actual, err := n.filesToDelete(labels)
		if err != nil {
			t.Errorf("Did not expect error got: %s", err)
		}
		sort.Strings(expected)
		sort.Strings(actual)
		if !reflect.DeepEqual(expected, actual) {
			t.Errorf("Expected files: %v, got: %v", expected, actual)
		}
	})
	t.Run("Works When there are multiple files to delete", func(t *testing.T) {
		n := &nodeLabelsToFiles{config: &config{directory: "test/resources"}}
		labels := map[string]string{"baz": ""}
		expected := []string{"test/resources/bar", "test/resources/foo"}
		actual, err := n.filesToDelete(labels)
		if err != nil {
			t.Errorf("Did not expect error got: %s", err)
		}
		sort.Strings(expected)
		sort.Strings(actual)
		if !reflect.DeepEqual(expected, actual) {
			t.Errorf("Expected files: %v, got: %v", expected, actual)
		}
	})
	t.Run("Works on directories", func(t *testing.T) {
		n := &nodeLabelsToFiles{config: &config{directory: "test"}}
		labels := map[string]string{"baz": ""}
		expected := []string{"test/resources", "test/resources/bar",
			"test/resources/foo"}
		actual, err := n.filesToDelete(labels)
		if err != nil {
			t.Errorf("Did not expect error got: %s", err)
		}
		sort.Strings(expected)
		sort.Strings(actual)
		if !reflect.DeepEqual(expected, actual) {
			t.Errorf("Expected files: %v, got: %v", expected, actual)
		}
	})
	t.Run("Works when labels includes paths directories", func(t *testing.T) {
		n := &nodeLabelsToFiles{config: &config{directory: "test"}}
		labels := map[string]string{"baz": "", "resources/bar": ""}
		expected := []string{"test/resources/foo"}
		actual, err := n.filesToDelete(labels)
		if err != nil {
			t.Errorf("Did not expect error got: %s", err)
		}
		sort.Strings(expected)
		sort.Strings(actual)
		if !reflect.DeepEqual(expected, actual) {
			t.Errorf("Expected files: %v, got: %v", expected, actual)
		}
	})
	t.Run("Works when configured directory ends with a /", func(t *testing.T) {
		n := &nodeLabelsToFiles{config: &config{directory: "test/"}}
		labels := map[string]string{"baz": "", "resources/bar": ""}
		expected := []string{"test/resources/foo"}
		actual, err := n.filesToDelete(labels)
		if err != nil {
			t.Errorf("Did not expect error got: %s", err)
		}
		sort.Strings(expected)
		sort.Strings(actual)
		if !reflect.DeepEqual(expected, actual) {
			t.Errorf("Expected files: %v, got: %v", expected, actual)
		}
	})
}

func TestValdate(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		config := &config{mode: Always, nodeName: "aRealNode", directory: "test"}
		err := config.Validate()
		if err != nil {
			t.Errorf("Expected no error got: %s", err)
		}
	})
	t.Run("Directory Not Provided", func(t *testing.T) {
		config := &config{mode: Always, nodeName: "aRealNode"}
		err := config.Validate()
		expectedError := "directory is not configured"
		if err == nil || err.Error() != expectedError {
			t.Errorf("Expected: %s, got: %v", expectedError, err)
		}
	})
	t.Run("nodeName Not Provided", func(t *testing.T) {
		config := &config{mode: Always, directory: "test"}
		err := config.Validate()
		expectedError := "nodename is not configured"
		if err == nil || err.Error() != expectedError {
			t.Errorf("Expected: %s, got: %v", expectedError, err)
		}
	})
	t.Run("mode Not Provided", func(t *testing.T) {
		config := &config{nodeName: "foo", directory: "test"}
		err := config.Validate()
		expectedError := "mode is not configured"
		if err == nil || err.Error() != expectedError {
			t.Errorf("Expected: %s, got: %v", expectedError, err)
		}
	})
	t.Run("wrong mode provided", func(t *testing.T) {
		config := &config{mode: "Wrong", nodeName: "foo", directory: "test"}
		err := config.Validate()
		expectedError := "mode should be one of once or always"
		if err == nil || err.Error() != expectedError {
			t.Errorf("Expected: %s, got: %v", expectedError, err)
		}
	})
}
