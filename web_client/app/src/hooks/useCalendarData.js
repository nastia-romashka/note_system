import { useEffect, useMemo, useState } from "react";

import { createNote, fetchCalendarNotes } from "../api/notesApi";
import { PREVIEW_DATA } from "../preview/previewData";

export function useCalendarData({ token, workspaceId, categories, uiPreview, setMessage, setLoading, onOpenNote }) {
  const [currentMonth, setCurrentMonth] = useState(startOfMonth(new Date()));
  const [selectedDay, setSelectedDay] = useState(startOfDay(new Date()));
  const [calendarNotes, setCalendarNotes] = useState(uiPreview ? collectPreviewCalendarNotes() : []);
  const [createDialog, setCreateDialog] = useState(null);

  const monthRange = useMemo(() => getMonthRange(currentMonth), [currentMonth]);

  useEffect(() => {
    if (uiPreview || !token) {
      return;
    }

    void loadMonth(monthRange.from, monthRange.to);
  }, [monthRange.from, monthRange.to, token, workspaceId, uiPreview]);

  const notesByDay = useMemo(() => {
    const buckets = new Map();

    for (const note of calendarNotes) {
      const startAt = note?.event?.start_at;
      if (!startAt) {
        continue;
      }

      const key = formatDayKey(new Date(startAt * 1000));
      const notes = buckets.get(key) || [];
      notes.push(note);
      buckets.set(key, notes);
    }

    for (const [key, notes] of buckets.entries()) {
      notes.sort((left, right) => (left?.event?.start_at || 0) - (right?.event?.start_at || 0));
      buckets.set(key, notes);
    }

    return buckets;
  }, [calendarNotes]);

  const selectedDayNotes = notesByDay.get(formatDayKey(selectedDay)) || [];

  async function loadMonth(from, to) {
    try {
      setLoading(true);
      const notes = await fetchCalendarNotes(token, { from, to, workspaceId });
      setCalendarNotes(Array.isArray(notes) ? notes : []);
    } catch (error) {
      handleError(error);
    } finally {
      setLoading(false);
    }
  }

  function openCreateDialog() {
    setCreateDialog({
      categoryUuid: firstCategoryUuid(categories),
      header: "",
      date: formatDateInput(selectedDay),
      startTime: "09:00",
      endTime: "10:00",
    });
  }

  function closeCreateDialog() {
    setCreateDialog(null);
  }

  async function confirmCreateFromCalendar() {
    if (!createDialog) {
      return;
    }

    const header = createDialog.header.trim();
    if (!createDialog.categoryUuid) {
      setMessage({ type: "warning", text: "Выберите категорию для заметки." });
      return;
    }
    if (!header) {
      setMessage({ type: "warning", text: "Введите заголовок заметки." });
      return;
    }

    const event = buildEventFromSchedule({
      date: createDialog.date,
      startTime: createDialog.startTime,
      endTime: createDialog.endTime,
      enabled: true,
    });
    if (!event) {
      setMessage({ type: "warning", text: "Укажите дату и время заметки." });
      return;
    }

    if (uiPreview) {
      const now = Math.floor(Date.now() / 1000);
      const previewNote = {
        uuid: `preview-${Date.now()}`,
        header,
        body: "",
        short_body: "",
        category_uuid: createDialog.categoryUuid,
        created_date: now,
        updated_at: now,
        tags: [],
        event,
      };
      setCalendarNotes((current) => [...current, previewNote]);
      closeCreateDialog();
      openNote(previewNote);
      return;
    }

    try {
      setLoading(true);
      const created = await createNote(token, {
        header,
        body: "",
        category_uuid: createDialog.categoryUuid,
        tags: [],
        event,
      }, workspaceId);
      closeCreateDialog();
      await loadMonth(monthRange.from, monthRange.to);
      setMessage({ type: "success", text: "Заметка создана и добавлена в календарь." });
      openNote(created);
    } catch (error) {
      handleError(error);
    } finally {
      setLoading(false);
    }
  }

  function openNote(note) {
    if (!note?.uuid) {
      return;
    }

    onOpenNote?.(note);
  }

  function handleError(error) {
    setMessage({
      type: "error",
      text: error instanceof Error ? error.message : "Произошла ошибка.",
    });
  }

  return {
    currentMonth,
    setCurrentMonth,
    selectedDay,
    setSelectedDay,
    selectedDayNotes,
    notesByDay,
    createDialog,
    setCreateDialog,
    openCreateDialog,
    closeCreateDialog,
    confirmCreateFromCalendar,
    openNote,
  };
}

function collectPreviewCalendarNotes() {
  return Object.values(PREVIEW_DATA.notes)
    .flatMap((items) => items || [])
    .filter((note) => note?.event?.enabled);
}

function firstCategoryUuid(categories) {
  for (const category of categories || []) {
    if (category?.uuid) {
      return category.uuid;
    }

    const nested = firstCategoryUuid(category?.children || []);
    if (nested) {
      return nested;
    }
  }

  return "";
}

function startOfMonth(date) {
  return new Date(date.getFullYear(), date.getMonth(), 1);
}

function startOfDay(date) {
  return new Date(date.getFullYear(), date.getMonth(), date.getDate());
}

function endOfDay(date) {
  return new Date(date.getFullYear(), date.getMonth(), date.getDate(), 23, 59, 59, 999);
}

function getMonthRange(date) {
  const monthStart = startOfMonth(date);
  const monthEnd = endOfDay(new Date(date.getFullYear(), date.getMonth() + 1, 0));

  return {
    from: Math.floor(monthStart.getTime() / 1000),
    to: Math.floor(monthEnd.getTime() / 1000),
  };
}

function formatDayKey(date) {
  return formatDateInput(date);
}

function formatDateInput(date) {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

function buildEventFromSchedule({ date, startTime, endTime, enabled }) {
  if (!date || !startTime || !endTime) {
    return null;
  }

  const startDate = new Date(`${date}T${startTime}:00`);
  const endDate = new Date(`${date}T${endTime}:00`);
  if (Number.isNaN(startDate.getTime()) || Number.isNaN(endDate.getTime())) {
    return null;
  }
  if (endDate.getTime() < startDate.getTime()) {
    return null;
  }

  return {
    enabled,
    start_at: Math.floor(startDate.getTime() / 1000),
    end_at: Math.floor(endDate.getTime() / 1000),
  };
}
