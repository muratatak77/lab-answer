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

package repo_test

import (
	"context"
	"testing"
	"time"

	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/repo/collection"
	"github.com/apache/answer/internal/repo/unique"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_collectionRepo_UpdateCollectionNote(t *testing.T) {
	ctx := context.TODO()
	collectionRepo := collection.NewCollectionRepo(testDataSource, unique.NewUniqueIDRepo(testDataSource))

	first := &entity.Collection{UserID: "9100", ObjectID: "10010000000009101", UserCollectionGroupID: "1"}
	require.NoError(t, collectionRepo.AddCollection(ctx, first))
	second := &entity.Collection{UserID: "9100", ObjectID: "10010000000009102", UserCollectionGroupID: "1"}
	require.NoError(t, collectionRepo.AddCollection(ctx, second))
	defer func() {
		_ = collectionRepo.RemoveCollection(ctx, first.ID)
		_ = collectionRepo.RemoveCollection(ctx, second.ID)
	}()

	// make the order of the bookmark list deterministic: second is the latest one
	_, err := testDataSource.DB.Context(ctx).ID(first.ID).Cols("updated_at").NoAutoTime().
		Update(&entity.Collection{UpdatedAt: time.Now().Add(-time.Hour)})
	require.NoError(t, err)

	before, exist, err := collectionRepo.GetOneByObjectIDAndUser(ctx, first.UserID, first.ObjectID)
	require.NoError(t, err)
	require.True(t, exist)
	assert.Empty(t, before.Note)

	require.NoError(t, collectionRepo.UpdateCollectionNote(ctx, first.ID, "compare with the accepted answer"))

	after, exist, err := collectionRepo.GetOneByObjectIDAndUser(ctx, first.UserID, first.ObjectID)
	require.NoError(t, err)
	require.True(t, exist)
	assert.Equal(t, "compare with the accepted answer", after.Note)
	assert.True(t, before.UpdatedAt.Equal(after.UpdatedAt), "updating a note must not change updated_at")

	list, total, err := collectionRepo.SearchList(ctx, &entity.CollectionSearch{
		Collection: entity.Collection{UserID: first.UserID},
		Page:       1,
		PageSize:   10,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	require.Len(t, list, 2)
	assert.Equal(t, second.ID, list[0].ID)
	assert.Empty(t, list[0].Note)
	assert.Equal(t, first.ID, list[1].ID)
	assert.Equal(t, "compare with the accepted answer", list[1].Note)

	require.NoError(t, collectionRepo.UpdateCollectionNote(ctx, first.ID, ""))
	cleared, _, err := collectionRepo.GetOneByObjectIDAndUser(ctx, first.UserID, first.ObjectID)
	require.NoError(t, err)
	assert.Empty(t, cleared.Note)
}
