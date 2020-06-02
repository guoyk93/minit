package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRemoveRotationMark(t *testing.T) {
	assert.Equal(t, "hello.log", rotationMarkRemove("hello.ROT2020-02-32.log"))
	assert.Equal(t, "hello.log", rotationMarkRemove("hello.ROT2020022.log"))
	assert.Equal(t, "hello.log", rotationMarkRemove("hello.ROT202-022.log"))
}

func TestWildcardRotationFile(t *testing.T) {
	assert.Equal(t, "hello.ROT*.log", rotationMarkAdd("hello.log", "*"))
	assert.Equal(t, "hello.ROT*", rotationMarkAdd("hello", "*"))
	assert.Equal(t, ".ROT*.hello", rotationMarkAdd(".hello", "*"))
}

func TestRotationMarkPattern(t *testing.T) {
	subs := rotationMarkPattern.FindStringSubmatch("hello.ROT111.log")
	t.Log(subs)
}

func TestRotationMarkAdd(t *testing.T) {
	assert.Equal(t, "hello.ROT000000000011.log", rotationMarkAdd("hello.log", fmt.Sprintf("%012d", 11)))
}
