package factlist_test

import (
	"context"
	"testing"

	"github.com/markhaur/trivia/pkg"
	"github.com/markhaur/trivia/pkg/factlist"
	"github.com/markhaur/trivia/pkg/inmem"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSave(t *testing.T) {
	var (
		assert = require.New(t)
		svc    = factlist.NewService(inmem.NewFactRepository())
	)

	fact := pkg.Fact{Question: "what is your github username?", Answer: "markhaur"}
	savedFact, err := svc.Save(context.TODO(), fact)

	assert.NoError(err, "could not save task")
	assert.NotZero(savedFact.ID)
	assert.Positive(savedFact.ID)
	assert.Equal(savedFact.Question, fact.Question)
	assert.Equal(savedFact.Answer, fact.Answer)
	assert.NotNil(savedFact.CreatedAt)
}

func TestList(t *testing.T) {
	var (
		require = require.New(t)
		assert  = assert.New(t)
		svc     = factlist.NewService(inmem.NewFactRepository())
	)

	expected := []pkg.Fact{
		{Question: "what is your github username?", Answer: "markhaur"},
		{Question: "what's your favourite language?", Answer: "Go"},
		{Question: "what's the name of current project?", Answer: "trivia"},
	}

	for _, fact := range expected {
		_, err := svc.Save(context.TODO(), fact)
		require.NoError(err, "could not save fact")
	}

	list, err := svc.List(context.TODO())
	require.NoError(err, "could not list facts")

	for i := range list {
		assert.Positive(list[i].ID)
		assert.NotZero(list[i].ID)
		assert.Equal(expected[i].Question, list[i].Question)
		assert.Equal(expected[i].Answer, list[i].Answer)
	}
}

func TestRemove(t *testing.T) {
	var (
		require = require.New(t)
		assert  = assert.New(t)
		svc     = factlist.NewService(inmem.NewFactRepository())
	)

	fact, err := svc.Save(context.TODO(), pkg.Fact{Question: "what is your github username?", Answer: "markhaur"})
	require.NoError(err, "could not save fact")

	require.NoError(svc.Remove(context.TODO(), fact.ID), "could not remove fact")

	list, err := svc.List(context.TODO())
	assert.NoError(err, "could not list facts")
	assert.Empty(list, "expected list to be empty after removing fact")
}

func TestUpdate(t *testing.T) {
	var (
		require = require.New(t)
		assert  = assert.New(t)
		svc     = factlist.NewService(inmem.NewFactRepository())
	)

	fact, err := svc.Save(context.TODO(), pkg.Fact{Question: "what is your github username?", Answer: "I don't know"})
	require.NoError(err, "could not save fact")

	fact.Question = "have you figured out your username?"
	fact.Answer = "yeah yeah, it's markhaur :p"

	fact, isCreated, err := svc.Update(context.TODO(), *fact)
	require.NoError(err, "could not update fact")
	assert.False(isCreated)
	assert.NotNil(fact)

	list, err := svc.List(context.TODO())
	assert.NoError(err, "could not list facts")
	assert.Equal(1, len(list), "unexpected number of tasks after update")
	assert.Equal(list[0].ID, fact.ID, "expected IDs to match")
	assert.Equal(list[0].Question, fact.Question, "expected Question to be updated")
	assert.Equal(list[0].Answer, fact.Answer, "expected Answer to be updated")
	assert.Equal(list[0].CreatedAt, fact.CreatedAt, "expected CreatedAt to be match")
}
