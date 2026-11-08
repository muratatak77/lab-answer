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

package migrations

import (
	"context"
	"testing"
	"time"

	"github.com/apache/answer/internal/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

// collectionBeforeV36 is the collection table before the note column was added.
type collectionBeforeV36 struct {
	ID                    string    `xorm:"not null pk default 0 BIGINT(20) id"`
	CreatedAt             time.Time `xorm:"created not null default CURRENT_TIMESTAMP TIMESTAMP created_at"`
	UpdatedAt             time.Time `xorm:"updated not null default CURRENT_TIMESTAMP TIMESTAMP updated_at"`
	UserID                string    `xorm:"not null default 0 BIGINT(20) INDEX user_id"`
	ObjectID              string    `xorm:"not null default 0 BIGINT(20) object_id"`
	UserCollectionGroupID string    `xorm:"not null default 0 BIGINT(20) user_collection_group_id"`
}

func (collectionBeforeV36) TableName() string {
	return "collection"
}

func TestAddCollectionNote(t *testing.T) {
	x, err := xorm.NewEngine("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() {
		_ = x.Close()
	}()
	require.NoError(t, x.Sync(new(collectionBeforeV36)))

	_, err = x.Insert(&collectionBeforeV36{
		ID:                    "10010000000000001",
		UserID:                "1",
		ObjectID:              "10010000000000002",
		UserCollectionGroupID: "1",
	})
	require.NoError(t, err)

	var noteMigration Migration
	for _, m := range GetMigrations() {
		if m.Version() == "v2.0.5" {
			noteMigration = m
			break
		}
	}
	require.NotNil(t, noteMigration)
	require.NoError(t, noteMigration.Migrate(context.Background(), x))

	tables, err := x.DBMetas()
	require.NoError(t, err)
	var columns []string
	for _, table := range tables {
		if table.Name == "collection" {
			columns = table.ColumnsSeq()
		}
	}
	assert.Contains(t, columns, "note")

	existing := &entity.Collection{}
	exist, err := x.ID("10010000000000001").Get(existing)
	require.NoError(t, err)
	require.True(t, exist)
	assert.Equal(t, "10010000000000002", existing.ObjectID)
	assert.Empty(t, existing.Note)

	existing.Note = "read later"
	_, err = x.ID(existing.ID).Cols("note").Update(existing)
	require.NoError(t, err)

	// running the migration again must be harmless
	require.NoError(t, addCollectionNote(context.Background(), x))
	updated := &entity.Collection{}
	_, err = x.ID(existing.ID).Get(updated)
	require.NoError(t, err)
	assert.Equal(t, "read later", updated.Note)
}
