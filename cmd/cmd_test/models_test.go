package cmd_test

import (
	"reflect"
	"testing"

	"github.com/2KAbhishek/gh-repo-man/cmd"
)

func TestTopicNames(t *testing.T) {
	repo := cmd.Repo{
		Topics: []cmd.Topic{
			{Name: "go"},
			{Name: "cli"},
			{Name: "github"},
		},
	}

	expected := []string{"go", "cli", "github"}
	result := repo.TopicNames()

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("TopicNames() = %v, want %v", result, expected)
	}
}

func TestTopicNamesEmpty(t *testing.T) {
	repo := cmd.Repo{
		Topics: []cmd.Topic{},
	}

	expected := []string{}
	result := repo.TopicNames()

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("TopicNames() with empty topics = %v, want %v", result, expected)
	}
}
