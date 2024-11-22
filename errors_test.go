package errors_test

import (
	"context"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dohernandez/errors"
)

func TestNew(t *testing.T) {
	t.Parallel()

	err := errors.New("failed")
	require.Error(t, err, "it is not an error")

	expected := "failed"
	assert.EqualError(t, err, expected, "error message mismatch, got %s want %s", err, expected)
}

func TestNewf(t *testing.T) {
	t.Parallel()

	err := errors.Newf("oops: %v", "failed")
	require.Error(t, err, "it is not an error")

	expected := "oops: failed"
	assert.EqualError(t, err, expected, "error message mismatch, got %s want %s", err, expected)
}

func TestWrap(t *testing.T) {
	t.Parallel()

	t.Run("Wrap with error", func(t *testing.T) {
		t.Parallel()

		err := errors.New("failed")

		errWrap := errors.Wrap(err, "oops")
		require.Error(t, errWrap, "it is not an error")

		expected := "oops: failed"
		assert.EqualError(t, errWrap, expected, "error message mismatch, got %s want %s", errWrap, expected)
	})

	t.Run("Wrap with nil", func(t *testing.T) {
		t.Parallel()

		errWrap := errors.Wrap(nil, "oops")
		require.NoError(t, errWrap, "error should be nil")
	})
}

func TestWrapf(t *testing.T) {
	t.Parallel()

	t.Run("Wrapf with error", func(t *testing.T) {
		t.Parallel()

		err := errors.New("failed")

		errWrap := errors.Wrapf(err, "oops id %d", 5)
		require.Error(t, errWrap, "it is not an error")

		expected := "oops id 5: failed"
		assert.EqualError(t, errWrap, expected, "error message mismatch, got %s want %s", errWrap, expected)
	})

	t.Run("Wrapf with nil", func(t *testing.T) {
		t.Parallel()

		errWrap := errors.Wrapf(nil, "oops id %d", 5)
		require.NoError(t, errWrap, "error should be nil")
	})
}

func TestWrapError(t *testing.T) {
	t.Parallel()

	t.Run("WrapWithError for errors", func(t *testing.T) {
		t.Parallel()

		err := errors.New("failed")
		sErr := errors.New("oops")

		errWrap := errors.WrapError(err, sErr)
		require.Error(t, errWrap, "it is not an error")

		expected := "oops: failed"
		assert.EqualError(t, errWrap, expected, "error message mismatch, got %s want %s", errWrap, expected)
	})

	t.Run("WrapWithError with cause nil", func(t *testing.T) {
		t.Parallel()

		sErr := errors.New("oops")

		errWrap := errors.WrapError(nil, sErr)
		require.Error(t, errWrap, "it is not an error")

		expected := "oops"
		require.EqualError(t, errWrap, expected, "error message mismatch, got %s want %s", errWrap, expected)

		require.Equal(t, sErr, errWrap)
	})

	t.Run("WrapWithError with supplied nil", func(t *testing.T) {
		t.Parallel()

		err := errors.New("oops")

		errWrap := errors.WrapError(err, nil)
		require.Error(t, errWrap, "it is not an error")

		expected := "oops"
		require.EqualError(t, errWrap, expected, "error message mismatch, got %s want %s", errWrap, expected)

		require.Equal(t, err, errWrap)
	})
}

type enrichedError interface {
	Tuples() []interface{}
	Fields() map[string]interface{}
}

func TestEnriched(t *testing.T) {
	t.Parallel()

	t.Run("Enrich error", func(t *testing.T) {
		t.Parallel()

		err := errors.New("failed")

		errEnriched := errors.Enrich(err, "id", 5)
		require.Error(t, errEnriched, "it is not an error")

		expected := "failed"
		require.EqualError(t, errEnriched, expected, "error message mismatch, got %s want %s", errEnriched, expected)

		errKV, ok := errEnriched.(enrichedError)
		require.True(t, ok, "error does not implement enrichedError interface")
		require.Equal(t, []interface{}{"id", 5}, errKV.Tuples())
	})

	t.Run("Enrich error and wrap", func(t *testing.T) {
		t.Parallel()

		err := errors.New("failed")

		errEnriched := errors.Enrich(err, "id", 5)
		require.Error(t, errEnriched, "it is not an error")

		sErr := errors.WrapError(errors.New("oops"), errEnriched)

		expected := "failed: oops"
		require.EqualError(t, sErr, expected, "error message mismatch, got %s want %s", sErr, expected)

		errKV, ok := errEnriched.(enrichedError)
		require.True(t, ok, "error does not implement enrichedError interface")
		require.Equal(t, []interface{}{"id", 5}, errKV.Tuples())
	})

	t.Run("Enrich error, malformed", func(t *testing.T) {
		t.Parallel()

		err := errors.New("failed")

		errEnriched := errors.Enrich(err, "id", "5", 5)
		require.Error(t, errEnriched, "it is not an error")

		expected := "failed"
		require.EqualError(t, errEnriched, expected, "error message mismatch, got %s want %s", errEnriched, expected)

		_, ok := errEnriched.(enrichedError)
		require.False(t, ok, "error does implement enrichedError interface")
	})

	t.Run("EnrichWrapWithError error", func(t *testing.T) {
		t.Parallel()

		err := errors.New("failed")
		sErr := errors.New("oops")

		errEnrichedWrap := errors.EnrichWrapError(err, sErr, "id", 5)
		require.Error(t, errEnrichedWrap, "it is not an error")

		require.ErrorIs(t, errEnrichedWrap, sErr)
		require.ErrorIs(t, errEnrichedWrap, err)

		expected := "oops: failed"
		require.EqualError(t, errEnrichedWrap, expected, "error message mismatch, got %s want %s", errEnrichedWrap, expected)

		errKV, ok := errEnrichedWrap.(enrichedError)
		require.True(t, ok, "error does not implement enrichedError interface")
		require.Equal(t, []interface{}{"id", 5}, errKV.Tuples())
	})

	t.Run("Enrich enriched error", func(t *testing.T) {
		t.Parallel()

		err := errors.New("failed")

		errEnriched := errors.Enrich(err, "id", 5)
		require.Error(t, errEnriched, "it is not an error")

		errEnriched2 := errors.Enrich(errEnriched, "number", 6, "hash", "0X0")
		require.Error(t, errEnriched2, "it is not an error")

		expected := "failed"
		require.EqualError(t, errEnriched, expected, "error message mismatch, got %s want %s", errEnriched, expected)

		errKV, ok := errEnriched.(enrichedError)
		require.True(t, ok, "error does not implement enrichedError interface")
		require.Equal(t, []interface{}{"id", 5}, errKV.Tuples())

		expected2 := "failed"
		require.EqualError(t, errEnriched2, expected2, "error message mismatch, got %s want %s", errEnriched, expected)

		errKV, ok = errEnriched2.(enrichedError)
		require.True(t, ok, "error does not implement enrichedError interface")
		require.Equal(t, []interface{}{"number", 6, "hash", "0X0", "id", 5}, errKV.Tuples())
	})

	t.Run("Enrich enriched cause error", func(t *testing.T) {
		t.Parallel()

		err := errors.New("failed")

		errEnriched := errors.Enrich(errors.Wrap(err, "stream blocks"), "block_hash", "0X0")
		require.Error(t, errEnriched)

		errEnriched2 := errors.EnrichWrapError(errors.New("oops"), errEnriched, "bInt", big.NewInt(42))
		require.Error(t, errEnriched2)

		expected := "stream blocks: failed"
		require.EqualError(t, errEnriched, expected, "error message mismatch, got %s want %s", errEnriched, expected)

		errKV, ok := errEnriched.(enrichedError)
		require.True(t, ok, "error does not implement enrichedError interface")
		require.Equal(t, []interface{}{"block_hash", "0X0"}, errKV.Tuples())

		expected2 := "stream blocks: failed: oops"
		require.EqualError(t, errEnriched2, expected2, "error message mismatch, got %s want %s", errEnriched, expected)

		errKV, ok = errEnriched2.(enrichedError)
		require.True(t, ok, "error does not implement enrichedError interface")
		require.Equal(t, []interface{}{"bInt", big.NewInt(42), "block_hash", "0X0"}, errKV.Tuples())
	})
}

func Test_Unwrap(t *testing.T) {
	t.Parallel()

	t.Run("Unwrap for errors.Wrap", func(t *testing.T) {
		t.Parallel()

		err := errors.New("failed")

		errWrap := errors.Wrap(err, "oops")
		require.Error(t, errWrap, "it is not an error")

		expected := "oops: failed"
		require.EqualError(t, errWrap, expected, "error message mismatch, got %s want %s", errWrap, expected)

		uErr := errors.Unwrap(errWrap)
		require.Error(t, uErr, "err does not implement Unwrap interface")

		expected = "failed"
		require.EqualError(t, uErr, expected, "error message mismatch, got %s want %s", uErr, expected)
	})

	t.Run("Unwrap for errors.Wrapf", func(t *testing.T) {
		t.Parallel()

		err := errors.New("failed")

		errWrap := errors.Wrapf(err, "oops id %d", 5)
		require.Error(t, errWrap, "it is not an error")

		expected := "oops id 5: failed"
		require.EqualError(t, errWrap, expected, "error message mismatch, got %s want %s", errWrap, expected)

		uErr := errors.Unwrap(errWrap)
		require.Error(t, uErr, "err does not implement Unwrap interface")

		expected = "failed"
		require.EqualError(t, uErr, expected, "error message mismatch, got %s want %s", uErr, expected)
	})

	t.Run("Unwrap for errors.WrapWithError", func(t *testing.T) {
		t.Parallel()

		err := errors.New("failed")
		sErr := errors.New("oops")

		errWrap := errors.WrapError(err, sErr)
		require.Error(t, errWrap, "it is not an error")

		uErr := errors.Unwrap(errWrap)
		require.Error(t, uErr, "err does not implement Unwrap interface")

		expected := "oops"
		require.EqualError(t, uErr, expected, "error message mismatch, got %s want %s", uErr, expected)
	})

	t.Run("Unwrap for errors.Enrich", func(t *testing.T) {
		t.Parallel()

		err := errors.New("failed")

		errEnriched := errors.Enrich(err, "id", 5)
		require.Error(t, errEnriched, "it is not an error")

		uErr := errors.Unwrap(errEnriched)
		require.Error(t, uErr, "err does not implement Unwrap interface")

		expected := "failed"
		require.EqualError(t, uErr, expected, "error message mismatch, got %s want %s", uErr, expected)
	})

	t.Run("Unwrap for errors.EnrichWrapWithError", func(t *testing.T) {
		t.Parallel()

		err := errors.New("failed")
		sErr := errors.New("oops")

		errEnriched := errors.EnrichWrapError(err, sErr, "id", 5)
		require.Error(t, errEnriched, "it is not an error")

		expected := "oops: failed"
		require.EqualError(t, errEnriched, expected, "error message mismatch, got %s want %s", errEnriched, expected)

		errKV, ok := errEnriched.(enrichedError)
		require.True(t, ok, "error does not implement enrichedError interface")
		require.Equal(t, []interface{}{"id", 5}, errKV.Tuples())

		errWrap := errors.Unwrap(errEnriched)
		require.Error(t, errWrap, "err does not implement Unwrap interface")

		expected = "oops: failed"
		require.EqualError(t, errWrap, expected, "error message mismatch, got %s want %s", errWrap, expected)

		uErr := errors.Unwrap(errWrap)
		require.Error(t, uErr, "err does not implement Unwrap interface")

		expected = "oops"
		require.EqualError(t, uErr, expected, "error message mismatch, got %s want %s", uErr, expected)
	})
}

func Test_Cause(t *testing.T) {
	t.Parallel()

	t.Run("Cause for errors.WrapWithError", func(t *testing.T) {
		t.Parallel()

		err := errors.New("failed")
		sErr := errors.New("oops")

		errWrap := errors.WrapError(err, sErr)
		require.Error(t, errWrap, "it is not an error")

		cErr := errors.Cause(errWrap)
		require.Error(t, cErr, "err does not implement Cause interface")

		expected := "failed"
		require.EqualError(t, cErr, expected, "error message mismatch, got %s want %s", cErr, expected)
	})
}

func Test_Is(t *testing.T) {
	t.Parallel()

	t.Run("Is for errors.New", func(t *testing.T) {
		t.Parallel()

		err := errors.New("failed")
		require.Error(t, err, "it is not an error")

		expected := errors.New("failed")
		require.ErrorIs(t, err, expected)
	})

	t.Run("no Is for errors.New", func(t *testing.T) {
		t.Parallel()

		err := errors.New("failed")
		require.Error(t, err, "it is not an error")

		require.NotErrorIs(t, err, context.Canceled)
	})

	t.Run("Is for errors.Newf", func(t *testing.T) {
		t.Parallel()

		err := errors.Newf("oops: %v", "failed")
		require.Error(t, err, "it is not an error")

		expected := errors.New("oops: failed")
		require.ErrorIs(t, err, expected)
	})

	t.Run("Is for errors.Wrap", func(t *testing.T) {
		t.Parallel()

		err := errors.New("failed")

		errWrap := errors.Wrap(err, "oops")
		require.Error(t, errWrap, "it is not an error")

		expected := errors.New("failed")
		require.ErrorIs(t, errWrap, expected)
	})

	t.Run("Is for errors.Wrapf", func(t *testing.T) {
		t.Parallel()

		err := errors.New("failed")

		errWrap := errors.Wrapf(err, "oops id %d", 5)
		require.Error(t, errWrap, "it is not an error")

		expected := errors.New("failed")
		require.ErrorIs(t, errWrap, expected)
	})

	t.Run("Is for errors.WrapWithError", func(t *testing.T) {
		t.Parallel()

		err := errors.New("failed")
		sErr := errors.New("oops")

		errWrap := errors.WrapError(err, sErr)
		require.Error(t, errWrap, "it is not an error")

		require.EqualError(t, errWrap, "oops: failed")

		require.ErrorIs(t, errWrap, err)

		require.ErrorIs(t, errWrap, sErr)
	})

	t.Run("no Is for errors.WrapWithError", func(t *testing.T) {
		t.Parallel()

		err := errors.New("failed")
		sErr := errors.New("oops")

		errWrap := errors.WrapError(err, sErr)
		require.Error(t, errWrap, "it is not an error")

		require.NotErrorIs(t, errWrap, context.Canceled)
	})

	t.Run("Is for errors.WrapWithError two levels", func(t *testing.T) {
		t.Parallel()

		sErr1 := errors.New("failed")
		sErr2 := errors.New("oops")

		errWrap := errors.WrapError(context.Canceled, sErr1)
		require.Error(t, errWrap, "it is not an error")

		errWrap = errors.WrapError(errWrap, sErr2)
		require.Error(t, errWrap, "it is not an error")

		require.ErrorIs(t, errWrap, sErr1)

		require.ErrorIs(t, errWrap, sErr2)

		require.ErrorIs(t, errWrap, context.Canceled)
	})

	t.Run("Is for errors.Enrich", func(t *testing.T) {
		t.Parallel()

		err := errors.New("failed")

		errEnrich := errors.Enrich(err, "id", 5)
		require.Error(t, errEnrich, "it is not an error")

		require.EqualError(t, errEnrich, "failed")

		errKV, ok := errEnrich.(enrichedError)
		require.True(t, ok, "error does not implement enrichedError interface")
		require.Equal(t, []interface{}{"id", 5}, errKV.Tuples())

		require.ErrorIs(t, errEnrich, err)
	})

	t.Run("no Is for errors.Enrich", func(t *testing.T) {
		t.Parallel()

		err := errors.New("failed")

		errEnrich := errors.Enrich(err, "id", 5)
		require.Error(t, errEnrich, "it is not an error")

		require.NotErrorIs(t, errEnrich, context.Canceled)
	})
}
