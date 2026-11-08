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
	"unicode/utf8"

	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	collectioncommon "github.com/apache/answer/internal/service/collection_common"
	questioncommon "github.com/apache/answer/internal/service/question_common"
	"github.com/apache/answer/pkg/converter"
	"github.com/segmentfault/pacman/errors"
)

// CollectionService user service
type CollectionService struct {
	collectionRepo      collectioncommon.CollectionRepo
	collectionGroupRepo CollectionGroupRepo
	questionCommon      *questioncommon.QuestionCommon
}

func NewCollectionService(
	collectionRepo collectioncommon.CollectionRepo,
	collectionGroupRepo CollectionGroupRepo,
	questionCommon *questioncommon.QuestionCommon,
) *CollectionService {
	return &CollectionService{
		collectionRepo:      collectionRepo,
		collectionGroupRepo: collectionGroupRepo,
		questionCommon:      questionCommon,
	}
}

func (cs *CollectionService) CollectionSwitch(ctx context.Context, req *schema.CollectionSwitchReq) (
	resp *schema.CollectionSwitchResp, err error) {
	collectionGroup, err := cs.collectionGroupRepo.CreateDefaultGroupIfNotExist(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	collection, exist, err := cs.collectionRepo.GetOneByObjectIDAndUser(ctx, req.UserID, req.ObjectID)
	if err != nil {
		return nil, err
	}
	if (!req.Bookmark && !exist) || (req.Bookmark && exist) {
		return nil, nil
	}

	if req.Bookmark {
		collection = &entity.Collection{
			UserID:                req.UserID,
			ObjectID:              req.ObjectID,
			UserCollectionGroupID: collectionGroup.ID,
		}
		err = cs.collectionRepo.AddCollection(ctx, collection)
	} else {
		err = cs.collectionRepo.RemoveCollection(ctx, collection.ID)
	}
	if err != nil {
		return nil, err
	}

	// For now, we only support bookmark for question, so we just update question collection count
	resp = &schema.CollectionSwitchResp{}
	resp.ObjectCollectionCount, err = cs.questionCommon.UpdateCollectionCount(ctx, req.ObjectID)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// maxCollectionNoteLength is the max length of a bookmark note in runes,
// it matches the VARCHAR(500) note column
const maxCollectionNoteLength = 500

// UpdateCollectionNote update the private note of the login user's bookmark
func (cs *CollectionService) UpdateCollectionNote(ctx context.Context, req *schema.UpdateCollectionNoteReq) (
	resp *schema.UpdateCollectionNoteResp, err error) {
	collection, exist, err := cs.collectionRepo.GetOneByObjectIDAndUser(ctx, req.UserID, req.ObjectID)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, errors.NotFound(reason.CollectionNotFound)
	}

	// the limit applies to the stored note, so it is checked after normalizing,
	// counted in runes like the validator's lte
	note := converter.NormalizePlainText(req.Note)
	if utf8.RuneCountInString(note) > maxCollectionNoteLength {
		return nil, errors.BadRequest(reason.CollectionNoteTooLong)
	}
	if note != collection.Note {
		if err = cs.collectionRepo.UpdateCollectionNote(ctx, collection.ID, note); err != nil {
			return nil, err
		}
	}
	return &schema.UpdateCollectionNoteResp{Note: note}, nil
}
