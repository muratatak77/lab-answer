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

package content

import (
	"context"
	"testing"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/pkg/uid"
)

// TestBuildPersonalCollectionList checks that PersonalCollectionPage returns each bookmark's
// private note as collection_note, in bookmark order, with both long and short question ids.
func TestBuildPersonalCollectionList(t *testing.T) {
	const (
		firstID   = "10010000000000001"
		secondID  = "10010000000000002"
		deletedID = "10010000000000003"
		missingID = "10010000000000004"
	)
	collections := []*entity.Collection{
		{ObjectID: secondID, Note: "second note"},
		{ObjectID: firstID, Note: ""},
		{ObjectID: missingID, Note: "question no longer found"},
		{ObjectID: deletedID, Note: "deleted note"},
	}

	tests := []struct {
		name    string
		shortID bool
	}{
		{name: "long ids", shortID: false},
		{name: "short ids", shortID: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := func(id string) string {
				if tt.shortID {
					return uid.EnShortID(id)
				}
				return id
			}
			if tt.shortID && key(firstID) == firstID {
				t.Fatalf("precondition failed: EnShortID(%q) is not a distinct short id", firstID)
			}
			// FindInfoByID keys the map by the formatted id, which is short when short ids are on.
			questionMaps := map[string]*schema.QuestionInfoResp{}
			for _, id := range []string{firstID, secondID, deletedID} {
				status := entity.QuestionStatusAvailable
				if id == deletedID {
					status = entity.QuestionStatusDeleted
				}
				questionMaps[key(id)] = &schema.QuestionInfoResp{
					ID: key(id), Title: "title " + id, Content: "content", HTML: "<p>content</p>",
					Status: status, UpdateUserInfo: &schema.UserBasicInfo{}, LastAnsweredUserInfo: &schema.UserBasicInfo{},
				}
			}
			ctx := context.WithValue(context.Background(), constant.ShortIDContextKey, tt.shortID)

			list := buildPersonalCollectionList(ctx, collections, questionMaps)

			want := []struct{ id, note, title string }{
				{key(secondID), "second note", "title " + secondID},
				{key(firstID), "", "title " + firstID},
				{key(deletedID), "deleted note", "Deleted question"},
			}
			if len(list) != len(want) {
				t.Fatalf("got %d items, want %d", len(list), len(want))
			}
			for i, w := range want {
				got := list[i]
				if got.ID != w.id || got.CollectionNote != w.note || got.Title != w.title {
					t.Errorf("item %d = {id %q, note %q, title %q}, want {id %q, note %q, title %q}",
						i, got.ID, got.CollectionNote, got.Title, w.id, w.note, w.title)
				}
				if got.Content != "" || got.HTML != "" || got.UpdateUserInfo != nil || got.LastAnsweredUserInfo != nil {
					t.Errorf("item %d: content, html and user infos should be cleared", i)
				}
			}
		})
	}
}
