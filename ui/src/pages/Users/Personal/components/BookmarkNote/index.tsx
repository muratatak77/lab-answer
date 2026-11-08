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

import { FC, FormEvent, KeyboardEvent, memo, useState } from 'react';
import { Button, Form } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';

import { Icon } from '@/components';
import { useToast } from '@/hooks';
import { updateBookmarkNote } from '@/services';

import './index.scss';

export const NOTE_MAX_LENGTH = 500;

// The server validates the note length in Unicode code points (Go runes),
// while String.length and the textarea maxLength count UTF-16 code units,
// so emoji and other astral characters would be counted twice here.
const noteLength = (text: string) => [...text].length;

const clampNote = (text: string) =>
  noteLength(text) > NOTE_MAX_LENGTH
    ? [...text].slice(0, NOTE_MAX_LENGTH).join('')
    : text;

interface Props {
  objectId: string;
  note?: string;
}

const Index: FC<Props> = ({ objectId, note = '' }) => {
  const { t } = useTranslation('translation', {
    keyPrefix: 'personal.bookmark_note',
  });
  const toast = useToast();
  const [savedNote, setSavedNote] = useState(note);
  const [draft, setDraft] = useState(note);
  const [editing, setEditing] = useState(false);
  const [saving, setSaving] = useState(false);
  const [errorMsg, setErrorMsg] = useState('');

  const startEdit = () => {
    setDraft(savedNote);
    setErrorMsg('');
    setEditing(true);
  };

  const cancelEdit = () => {
    setEditing(false);
    setErrorMsg('');
  };

  const handleSubmit = (evt?: FormEvent) => {
    evt?.preventDefault();
    if (saving) {
      return;
    }
    if (draft.trim() === savedNote) {
      setEditing(false);
      return;
    }
    setSaving(true);
    updateBookmarkNote({ object_id: objectId, note: draft })
      .then((resp) => {
        setSavedNote(resp?.note || '');
        setEditing(false);
        toast.onShow({
          msg: resp?.note ? t('saved') : t('removed'),
          variant: 'success',
        });
      })
      .catch((err) => {
        if (err?.isError) {
          const fieldError = err.list?.find((v) => v.error_field === 'note');
          setErrorMsg(fieldError?.error_msg || err.msg || '');
        }
      })
      .finally(() => {
        setSaving(false);
      });
  };

  const handleKeyDown = (evt: KeyboardEvent<HTMLTextAreaElement>) => {
    if (evt.key === 'Escape') {
      evt.preventDefault();
      cancelEdit();
    }
    if (evt.key === 'Enter' && (evt.metaKey || evt.ctrlKey)) {
      handleSubmit();
    }
  };

  if (editing) {
    return (
      <Form className="bookmark-note mt-2" onSubmit={handleSubmit}>
        <Form.Group controlId={`bookmark-note-${objectId}`}>
          <Form.Label className="visually-hidden">{t('label')}</Form.Label>
          <Form.Control
            as="textarea"
            rows={3}
            autoFocus
            value={draft}
            placeholder={t('placeholder')}
            isInvalid={!!errorMsg}
            disabled={saving}
            onChange={(e) => setDraft(clampNote(e.currentTarget.value))}
            onKeyDown={handleKeyDown}
          />
          <Form.Control.Feedback type="invalid">
            {errorMsg}
          </Form.Control.Feedback>
        </Form.Group>
        <div className="d-flex align-items-center mt-2">
          <Button type="submit" size="sm" variant="primary" disabled={saving}>
            {t('save')}
          </Button>
          <Button
            size="sm"
            variant="link"
            className="ms-2"
            disabled={saving}
            onClick={cancelEdit}>
            {t('cancel')}
          </Button>
          <span className="ms-auto small text-secondary">
            {t('length', { count: noteLength(draft), max: NOTE_MAX_LENGTH })}
          </span>
        </div>
      </Form>
    );
  }

  if (!savedNote) {
    return (
      <Button
        size="sm"
        variant="link"
        className="link-secondary p-0 mt-2 small"
        onClick={startEdit}>
        <Icon name="journal-plus" className="me-1" />
        {t('add')}
      </Button>
    );
  }

  return (
    <div className="bookmark-note mt-2 p-2 rounded bg-body-tertiary small">
      <div className="d-flex align-items-start">
        <Icon name="journal-text" className="me-2 text-secondary" />
        <div className="bookmark-note-text flex-grow-1 text-break">
          {savedNote}
        </div>
        <Button
          size="sm"
          variant="link"
          className="link-secondary p-0 ms-2 small"
          onClick={startEdit}>
          {t('edit')}
        </Button>
      </div>
    </div>
  );
};

export default memo(Index);
