/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package collection

import (
	"context"
	"errors"
	"testing"

	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	collectioncommon "github.com/apache/answer/internal/service/collection_common"
	pacmanErrors "github.com/segmentfault/pacman/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeCollectionRepo only implements the methods used by UpdateCollectionNote,
// calling any other method panics because of the nil embedded interface.
type fakeCollectionRepo struct {
	collectioncommon.CollectionRepo

	collections map[string]*entity.Collection // key: userID + "/" + objectID
	getErr      error
	updateErr   error
	updateCalls int
}

func newFakeCollectionRepo(collections ...*entity.Collection) *fakeCollectionRepo {
	repo := &fakeCollectionRepo{collections: make(map[string]*entity.Collection)}
	for _, c := range collections {
		repo.collections[c.UserID+"/"+c.ObjectID] = c
	}
	return repo
}

func (f *fakeCollectionRepo) GetOneByObjectIDAndUser(_ context.Context, userID string, objectID string) (
	*entity.Collection, bool, error) {
	if f.getErr != nil {
		return nil, false, f.getErr
	}
	c, ok := f.collections[userID+"/"+objectID]
	return c, ok, nil
}

func (f *fakeCollectionRepo) UpdateCollectionNote(_ context.Context, id string, note string) error {
	f.updateCalls++
	if f.updateErr != nil {
		return f.updateErr
	}
	for _, c := range f.collections {
		if c.ID == id {
			c.Note = note
		}
	}
	return nil
}

func TestCollectionService_UpdateCollectionNote(t *testing.T) {
	ctx := context.Background()

	t.Run("save normalized note", func(t *testing.T) {
		bookmark := &entity.Collection{ID: "10", UserID: "1", ObjectID: "100"}
		repo := newFakeCollectionRepo(bookmark)
		svc := NewCollectionService(repo, nil, nil)

		resp, err := svc.UpdateCollectionNote(ctx, &schema.UpdateCollectionNoteReq{
			ObjectID: "100",
			UserID:   "1",
			Note:     "  check the second answer\r\n\r\n\r\nit has a benchmark  ",
		})
		require.NoError(t, err)
		assert.Equal(t, "check the second answer\n\nit has a benchmark", resp.Note)
		assert.Equal(t, resp.Note, bookmark.Note)
		assert.Equal(t, 1, repo.updateCalls)
	})

	t.Run("empty note clears the existing note", func(t *testing.T) {
		bookmark := &entity.Collection{ID: "10", UserID: "1", ObjectID: "100", Note: "old note"}
		repo := newFakeCollectionRepo(bookmark)
		svc := NewCollectionService(repo, nil, nil)

		resp, err := svc.UpdateCollectionNote(ctx, &schema.UpdateCollectionNoteReq{
			ObjectID: "100",
			UserID:   "1",
			Note:     " \n ",
		})
		require.NoError(t, err)
		assert.Empty(t, resp.Note)
		assert.Empty(t, bookmark.Note)
		assert.Equal(t, 1, repo.updateCalls)
	})

	t.Run("unchanged note does not hit the database", func(t *testing.T) {
		bookmark := &entity.Collection{ID: "10", UserID: "1", ObjectID: "100", Note: "same"}
		repo := newFakeCollectionRepo(bookmark)
		svc := NewCollectionService(repo, nil, nil)

		resp, err := svc.UpdateCollectionNote(ctx, &schema.UpdateCollectionNoteReq{
			ObjectID: "100",
			UserID:   "1",
			Note:     "same ",
		})
		require.NoError(t, err)
		assert.Equal(t, "same", resp.Note)
		assert.Zero(t, repo.updateCalls)
	})

	t.Run("cannot update a bookmark of another user", func(t *testing.T) {
		bookmark := &entity.Collection{ID: "10", UserID: "1", ObjectID: "100", Note: "mine"}
		repo := newFakeCollectionRepo(bookmark)
		svc := NewCollectionService(repo, nil, nil)

		resp, err := svc.UpdateCollectionNote(ctx, &schema.UpdateCollectionNoteReq{
			ObjectID: "100",
			UserID:   "2",
			Note:     "not yours",
		})
		assert.Nil(t, resp)
		require.Error(t, err)
		var e *pacmanErrors.Error
		require.ErrorAs(t, err, &e)
		assert.Equal(t, reason.CollectionNotFound, e.Reason)
		assert.Equal(t, "mine", bookmark.Note)
		assert.Zero(t, repo.updateCalls)
	})

	t.Run("repository errors are returned", func(t *testing.T) {
		repo := newFakeCollectionRepo(&entity.Collection{ID: "10", UserID: "1", ObjectID: "100"})
		repo.updateErr = errors.New("db is gone")
		svc := NewCollectionService(repo, nil, nil)

		_, err := svc.UpdateCollectionNote(ctx, &schema.UpdateCollectionNoteReq{
			ObjectID: "100",
			UserID:   "1",
			Note:     "note",
		})
		require.EqualError(t, err, "db is gone")

		repo.getErr = errors.New("query failed")
		_, err = svc.UpdateCollectionNote(ctx, &schema.UpdateCollectionNoteReq{
			ObjectID: "100",
			UserID:   "1",
			Note:     "note",
		})
		assert.EqualError(t, err, "query failed")
	})
}
