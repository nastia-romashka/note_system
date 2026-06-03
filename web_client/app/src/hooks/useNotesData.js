import { useEffect, useMemo, useState } from "react";

import { createCategory, deleteCategory, fetchCategories, updateCategory } from "../api/categoriesApi";
import { createNote, deleteNote, duplicateNote, fetchNotes, fetchSearchNotes, updateNote } from "../api/notesApi";
import { createTag, fetchTags } from "../api/tagsApi";
import { PREVIEW_DATA } from "../preview/previewData";

export function useNotesData({ token, uiPreview, setMessage, setLoading }) {
  const [categories, setCategories] = useState(uiPreview ? PREVIEW_DATA.categories : []);
  const [notes, setNotes] = useState(uiPreview ? PREVIEW_DATA.notes["cat-1"] : []);
  const [tags, setTags] = useState(uiPreview ? PREVIEW_DATA.tags : []);
  const [searchResults, setSearchResults] = useState([]);

  const [selectedCategoryId, setSelectedCategoryId] = useState(uiPreview ? "cat-1" : "");
  const [selectedNoteId, setSelectedNoteId] = useState(uiPreview ? "note-1" : "");
  const [search, setSearch] = useState("");

  const [categoryForm, setCategoryForm] = useState({ name: "", color: "#9db8ff" });
  const [noteForm, setNoteForm] = useState({ header: "", body: "", tags: "" });
  const [noteEditorForm, setNoteEditorForm] = useState(createEmptyEditorForm());
  const [pendingCategoryDelete, setPendingCategoryDelete] = useState(null);
  const [subcategoryDraft, setSubcategoryDraft] = useState("");
  const [subcategoryParent, setSubcategoryParent] = useState(null);
  const [categoryEditor, setCategoryEditor] = useState(null);
  const [duplicateDialog, setDuplicateDialog] = useState(null);

  const filteredNotes = useMemo(() => {
    if (!uiPreview) {
      return notes;
    }

    const query = search.trim().toLowerCase();
    if (!query) {
      return notes;
    }

    return notes.filter((note) => {
      const haystack = `${note.header} ${note.body} ${note.short_body}`.toLowerCase();
      return haystack.includes(query);
    });
  }, [uiPreview, notes, search]);

  const selectedNote = useMemo(
    () =>
      filteredNotes.find((note) => note.uuid === selectedNoteId) ||
      notes.find((note) => note.uuid === selectedNoteId) ||
      null,
    [filteredNotes, notes, selectedNoteId],
  );

  const selectedCategory = useMemo(
    () => findCategoryById(categories, selectedCategoryId),
    [categories, selectedCategoryId],
  );

  const parsedEditorTags = useMemo(() => parseTagNames(noteEditorForm.tags), [noteEditorForm.tags]);

  useEffect(() => {
    if (uiPreview || !token) {
      return;
    }

    void bootstrap();
  }, [token]);

  useEffect(() => {
    if (uiPreview) {
      const nextNotes = PREVIEW_DATA.notes[selectedCategoryId] || [];
      setNotes(nextNotes);
      const nextNoteId = nextNotes.find((note) => note.uuid === selectedNoteId)?.uuid || nextNotes[0]?.uuid || "";
      setSelectedNoteId(nextNoteId);
      return;
    }

    if (!token || !selectedCategoryId) {
      setNotes([]);
      setSelectedNoteId("");
      return;
    }

    void loadNotes(selectedCategoryId, selectedNoteId);
  }, [selectedCategoryId, token]);

  useEffect(() => {
    if (uiPreview || !token) {
      return;
    }

    const query = search.trim();
    if (!query) {
      setSearchResults([]);
      return;
    }

    const timeoutId = window.setTimeout(() => {
      void loadSearchResults(query);
    }, 250);

    return () => window.clearTimeout(timeoutId);
  }, [search, token]);

  useEffect(() => {
    if (!selectedNote) {
      setNoteEditorForm(createEmptyEditorForm());
      return;
    }

    setNoteEditorForm({
      header: selectedNote.header || "",
      body: selectedNote.body || "",
      tags: stringifyTagNames(selectedNote.tags || [], tags),
      ...toEventEditorState(selectedNote.event),
    });
  }, [selectedNote, tags]);

  async function bootstrap(preferredCategoryId = selectedCategoryId, includeTags = true) {
    try {
      setLoading(true);
      setMessage(null);

      const [categoryList, tagList] = await Promise.all([
        fetchCategories(token),
        includeTags ? fetchTags(token) : Promise.resolve(tags),
      ]);

      setCategories(categoryList);
      setTags(tagList);

      const nextCategory = findCategoryById(categoryList, preferredCategoryId)?.uuid || categoryList[0]?.uuid || "";
      setSelectedCategoryId(nextCategory);

      if (!nextCategory) {
        setNotes([]);
        setSelectedNoteId("");
      }
    } catch (error) {
      handleError(error);
    } finally {
      setLoading(false);
    }
  }

  async function loadNotes(categoryId, preferredNoteId) {
    try {
      setLoading(true);
      const noteList = await fetchNotes(token, categoryId);
      applyNotesList(noteList, preferredNoteId);
    } catch (error) {
      handleError(error);
    } finally {
      setLoading(false);
    }
  }

  async function loadSearchResults(query) {
    try {
      const noteList = await fetchSearchNotes(token, { query });
      setSearchResults(Array.isArray(noteList) ? noteList : []);
    } catch (error) {
      setSearchResults([]);
      handleError(error);
    }
  }

  function applyNotesList(noteList, preferredNoteId) {
    const safeNoteList = Array.isArray(noteList) ? noteList : [];
    setNotes(safeNoteList);

    const nextNoteId = safeNoteList.some((note) => note.uuid === preferredNoteId)
      ? preferredNoteId
      : (safeNoteList[0]?.uuid || "");
    setSelectedNoteId(nextNoteId);
  }

  function handleSelectNote(noteId) {
    if (!noteId) {
      setSelectedNoteId("");
      return;
    }

    const note =
      notes.find((item) => item.uuid === noteId) ||
      filteredNotes.find((item) => item.uuid === noteId) ||
      null;

    if (note?.category_uuid && note.category_uuid !== selectedCategoryId) {
      setSelectedCategoryId(note.category_uuid);
    }

    setSelectedNoteId(noteId);
  }

  function handleSelectSearchResult(note) {
    if (!note?.uuid) {
      return;
    }

    handleOpenNote(note);
  }

  function handleOpenNote(note) {
    if (!note?.uuid) {
      return;
    }

    setSearch("");
    setSearchResults([]);

    if (note.category_uuid && note.category_uuid !== selectedCategoryId) {
      setSelectedCategoryId(note.category_uuid);
    }

    setSelectedNoteId(note.uuid);
  }

  function handleOpenGraphNode(node) {
    if (!node?.type || !node?.nodeId) {
      return;
    }

    setSearch("");
    setSearchResults([]);

    if (node.type === "category") {
      setSelectedCategoryId(node.nodeId);
      setSelectedNoteId("");
      return;
    }

    if (node.type === "note") {
      if (node.category_uuid) {
        setSelectedCategoryId(node.category_uuid);
      }
      setSelectedNoteId(node.nodeId);
    }
  }

  function openDuplicateDialog() {
    if (uiPreview) {
      setMessage({ type: "info", text: "Preview mode: дублирование заметок отключено." });
      return;
    }

    if (!selectedNote) {
      return;
    }

    setDuplicateDialog({
      sourceNoteUuid: selectedNote.uuid,
      categoryUuid: selectedNote.category_uuid || selectedCategoryId,
      header: `${selectedNote.header || "Без названия"} (копия)`,
      shortBody: selectedNote.short_body || selectedNote.body || "",
    });
  }

  function closeDuplicateDialog() {
    setDuplicateDialog(null);
  }

  async function confirmDuplicateNote() {
    if (!duplicateDialog?.sourceNoteUuid) {
      return;
    }

    const header = duplicateDialog.header.trim();
    if (!header) {
      setMessage({ type: "warning", text: "Введите заголовок копии заметки." });
      return;
    }
    if (!duplicateDialog.categoryUuid) {
      setMessage({ type: "warning", text: "Выберите категорию для копии заметки." });
      return;
    }

    try {
      setLoading(true);
      const duplicated = await duplicateNote(token, duplicateDialog.sourceNoteUuid, {
        category_uuid: duplicateDialog.categoryUuid,
        header,
      });

      setDuplicateDialog(null);
      setSelectedCategoryId(duplicated.category_uuid);
      await loadNotes(duplicated.category_uuid, duplicated.uuid);
      setMessage({ type: "success", text: "Заметка продублирована." });
    } catch (error) {
      handleError(error);
    } finally {
      setLoading(false);
    }
  }

  async function handleCreateCategory(event) {
    event.preventDefault();
    if (uiPreview) {
      setMessage({ type: "info", text: "Preview mode: создание категории отключено." });
      return;
    }

    try {
      setLoading(true);
      await createCategory(token, categoryForm);
      setCategoryForm({ name: "", color: randomCoolColor() });
      setMessage({ type: "success", text: "Категория создана." });
      await bootstrap(selectedCategoryId, false);
    } catch (error) {
      handleError(error);
    } finally {
      setLoading(false);
    }
  }

  function openSubcategoryDialog(category) {
    if (uiPreview) {
      setMessage({ type: "info", text: "Preview mode: создание подкатегорий отключено." });
      return;
    }

    setSubcategoryParent(category);
    setSubcategoryDraft("");
  }

  async function confirmCreateSubcategory() {
    if (!subcategoryParent || !subcategoryDraft.trim()) {
      return;
    }

    try {
      setLoading(true);
      await createCategory(token, {
        name: subcategoryDraft.trim(),
        color: randomCoolColor(),
        parent_uuid: subcategoryParent.uuid,
      });
      setMessage({ type: "success", text: `Подкатегория для "${subcategoryParent.name}" создана.` });
      setSubcategoryParent(null);
      setSubcategoryDraft("");
      await bootstrap(selectedCategoryId, false);
    } catch (error) {
      handleError(error);
    } finally {
      setLoading(false);
    }
  }

  function openDeleteCategoryDialog(categoryId) {
    if (uiPreview) {
      setMessage({ type: "info", text: "Preview mode: удаление категорий отключено." });
      return;
    }

    if (!categoryId) {
      return;
    }

    const category = findCategoryById(categories, categoryId);
    setPendingCategoryDelete(category || { uuid: categoryId, name: "Категория" });
  }

  function openCategoryEditor(category, mode = "menu") {
    if (!category?.uuid) {
      return;
    }

    setCategoryEditor({
      uuid: category.uuid,
      mode,
      name: category.name || "",
      color: category.color || "#9db8ff",
    });
  }

  function closeCategoryEditor() {
    setCategoryEditor(null);
  }

  function startRenameCategory(category) {
    if (uiPreview) {
      setMessage({ type: "info", text: "Preview mode: редактирование категорий отключено." });
      return;
    }

    openCategoryEditor(category, "rename");
  }

  function startRecolorCategory(category) {
    if (uiPreview) {
      setMessage({ type: "info", text: "Preview mode: редактирование категорий отключено." });
      return;
    }

    openCategoryEditor(category, "color");
  }

  async function submitCategoryRename() {
    if (!categoryEditor?.uuid) {
      return;
    }

    const name = categoryEditor.name.trim();
    if (!name) {
      setMessage({ type: "warning", text: "Введите название категории." });
      return;
    }

    try {
      setLoading(true);
      await updateCategory(token, categoryEditor.uuid, { name });
      setMessage({ type: "success", text: "Название категории обновлено." });
      await bootstrap(selectedCategoryId, false);
      setCategoryEditor(null);
    } catch (error) {
      handleError(error);
    } finally {
      setLoading(false);
    }
  }

  async function submitCategoryColor(color = categoryEditor?.color) {
    if (!categoryEditor?.uuid || !color) {
      return;
    }

    const name = categoryEditor.name.trim();
    if (!name) {
      setMessage({ type: "warning", text: "Введите название категории." });
      return;
    }

    try {
      setLoading(true);
      await updateCategory(token, categoryEditor.uuid, { name, color });
      setMessage({ type: "success", text: "Цвет категории обновлен." });
      await bootstrap(selectedCategoryId, false);
      setCategoryEditor((current) => (current ? { ...current, color, mode: "menu" } : current));
    } catch (error) {
      handleError(error);
    } finally {
      setLoading(false);
    }
  }

  async function confirmDeleteCategory() {
    if (!pendingCategoryDelete) {
      return;
    }

    const categoryId = pendingCategoryDelete.uuid;

    try {
      setLoading(true);
      await deleteCategory(token, categoryId);
      setMessage({ type: "success", text: "Категория удалена." });

      if (selectedCategoryId === categoryId) {
        setSelectedCategoryId("");
        setSelectedNoteId("");
        setNotes([]);
      }

      await bootstrap("", false);
    } catch (error) {
      handleError(error);
    } finally {
      setPendingCategoryDelete(null);
      setCategoryEditor(null);
      setLoading(false);
    }
  }

  async function ensureTagUUIDs(rawTags) {
    const tagNames = parseTagNames(rawTags);
    if (!tagNames.length) {
      return [];
    }

    let availableTags = tags;
    const existingNames = new Set(availableTags.map((tag) => tag.name.toLowerCase()));
    const missingNames = tagNames.filter((name) => !existingNames.has(name.toLowerCase()));

    for (const name of missingNames) {
      try {
        await createTag(token, name);
      } catch (error) {
        const errorMessage = error instanceof Error ? error.message.toLowerCase() : "";
        if (!errorMessage.includes("already exists")) {
          throw error;
        }
      }
    }

    if (missingNames.length > 0) {
      availableTags = await fetchTags(token);
      setTags(availableTags);
    }

    const expected = new Set(tagNames.map((name) => name.toLowerCase()));
    return availableTags.filter((tag) => expected.has(tag.name.toLowerCase())).map((tag) => tag.uuid);
  }

  async function handleCreateNote(event) {
    event.preventDefault();
    if (uiPreview) {
      setMessage({ type: "info", text: "Preview mode: создание заметок отключено." });
      return;
    }

    if (!selectedCategoryId) {
      setMessage({ type: "warning", text: "Сначала выберите категорию." });
      return;
    }

    const header = noteForm.header.trim();
    if (!header) {
      setMessage({ type: "warning", text: "Введите заголовок заметки." });
      return;
    }

    try {
      setLoading(true);
      const tagUUIDs = await ensureTagUUIDs(noteForm.tags);
      await createNote(token, {
        header,
        body: noteForm.body.trim(),
        category_uuid: selectedCategoryId,
        tags: tagUUIDs,
      });
      setNoteForm({ header: "", body: "", tags: "" });
      setMessage({ type: "success", text: "Заметка создана." });
      await loadNotes(selectedCategoryId, "");
    } catch (error) {
      handleError(error);
    } finally {
      setLoading(false);
    }
  }

  async function handleUpdateNote(event) {
    event.preventDefault();
    if (uiPreview) {
      setMessage({ type: "info", text: "Preview mode: редактирование заметок отключено." });
      return;
    }

    if (!selectedNote) {
      return;
    }

    const eventPayload = buildEventPayload(noteEditorForm);
    if (noteEditorForm.scheduled && !eventPayload) {
      setMessage({ type: "warning", text: "Укажите корректные дату и время события." });
      return;
    }

    try {
      setLoading(true);
      const tagUUIDs = await ensureTagUUIDs(noteEditorForm.tags);
      await updateNote(token, selectedNote.uuid, {
        header: noteEditorForm.header,
        body: noteEditorForm.body,
        tags: tagUUIDs,
        event: noteEditorForm.scheduled ? eventPayload : null,
      });
      setMessage({ type: "success", text: "Заметка обновлена." });
      await loadNotes(selectedCategoryId, selectedNote.uuid);
    } catch (error) {
      handleError(error);
    } finally {
      setLoading(false);
    }
  }

  async function handleDeleteNote(noteId) {
    if (uiPreview) {
      setMessage({ type: "info", text: "Preview mode: удаление заметок отключено." });
      return;
    }

    try {
      setLoading(true);
      await deleteNote(token, noteId);
      setMessage({ type: "success", text: "Заметка удалена." });
      await loadNotes(selectedCategoryId, "");
    } catch (error) {
      handleError(error);
    } finally {
      setLoading(false);
    }
  }

  function closeSubcategoryDialog() {
    setSubcategoryParent(null);
    setSubcategoryDraft("");
  }

  function resetNotesState() {
    setCategories([]);
    setNotes([]);
    setSearchResults([]);
    setTags([]);
    setCategoryEditor(null);
    setSelectedCategoryId("");
    setSelectedNoteId("");
    setSearch("");
  }

  function handleError(error) {
    setMessage({
      type: "error",
      text: error instanceof Error ? error.message : "Произошла ошибка.",
    });
  }

  return {
    categories,
    searchResults,
    filteredNotes,
    selectedNote,
    selectedCategory,
    selectedCategoryId,
    setSelectedCategoryId,
    selectedNoteId,
    setSelectedNoteId,
    handleSelectNote,
    handleSelectSearchResult,
    handleOpenNote,
    handleOpenGraphNode,
    search,
    setSearch,
    categoryForm,
    setCategoryForm,
    noteForm,
    setNoteForm,
    noteEditorForm,
    setNoteEditorForm,
    parsedEditorTags,
    pendingCategoryDelete,
    setPendingCategoryDelete,
    subcategoryDraft,
    setSubcategoryDraft,
    subcategoryParent,
    closeSubcategoryDialog,
    categoryEditor,
    duplicateDialog,
    setCategoryEditor,
    openCategoryEditor,
    closeCategoryEditor,
    startRenameCategory,
    startRecolorCategory,
    submitCategoryRename,
    submitCategoryColor,
    handleCreateCategory,
    openSubcategoryDialog,
    confirmCreateSubcategory,
    openDeleteCategoryDialog,
    confirmDeleteCategory,
    handleCreateNote,
    handleUpdateNote,
    handleDeleteNote,
    openDuplicateDialog,
    closeDuplicateDialog,
    confirmDuplicateNote,
    setDuplicateDialog,
    resetNotesState,
  };
}

function createEmptyEditorForm() {
  return {
    header: "",
    body: "",
    tags: "",
    ...emptyEventEditorState(),
  };
}

function emptyEventEditorState() {
  return {
    scheduled: false,
    eventEnabled: true,
    eventDate: "",
    eventStartTime: "09:00",
    eventEndTime: "10:00",
  };
}

function toEventEditorState(event) {
  if (!event?.start_at || !event?.end_at) {
    return emptyEventEditorState();
  }

  return {
    scheduled: true,
    eventEnabled: event.enabled !== false,
    eventDate: formatDateForInput(event.start_at),
    eventStartTime: formatTimeForInput(event.start_at),
    eventEndTime: formatTimeForInput(event.end_at),
  };
}

function buildEventPayload(form) {
  if (!form?.scheduled) {
    return null;
  }
  if (!form.eventDate || !form.eventStartTime || !form.eventEndTime) {
    return null;
  }

  const startDate = new Date(`${form.eventDate}T${form.eventStartTime}:00`);
  const endDate = new Date(`${form.eventDate}T${form.eventEndTime}:00`);
  if (Number.isNaN(startDate.getTime()) || Number.isNaN(endDate.getTime())) {
    return null;
  }
  if (endDate.getTime() < startDate.getTime()) {
    return null;
  }

  return {
    enabled: form.eventEnabled,
    start_at: Math.floor(startDate.getTime() / 1000),
    end_at: Math.floor(endDate.getTime() / 1000),
  };
}

function formatDateForInput(unixSeconds) {
  const date = new Date(unixSeconds * 1000);
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

function formatTimeForInput(unixSeconds) {
  const date = new Date(unixSeconds * 1000);
  const hours = String(date.getHours()).padStart(2, "0");
  const minutes = String(date.getMinutes()).padStart(2, "0");
  return `${hours}:${minutes}`;
}

function findCategoryById(categories, categoryId) {
  if (!categoryId) {
    return null;
  }

  for (const category of categories) {
    if (category.uuid === categoryId) {
      return category;
    }

    const children = Array.isArray(category.children) ? category.children : [];
    const nestedMatch = findCategoryById(children, categoryId);
    if (nestedMatch) {
      return nestedMatch;
    }
  }

  return null;
}

function parseTagNames(rawTags) {
  if (!rawTags.trim()) {
    return [];
  }

  const seen = new Set();

  return rawTags
    .split(",")
    .map((item) => item.trim())
    .filter(Boolean)
    .filter((item) => {
      const normalized = item.toLowerCase();
      if (seen.has(normalized)) {
        return false;
      }
      seen.add(normalized);
      return true;
    });
}

function stringifyTagNames(tagUUIDs, availableTags) {
  if (!tagUUIDs.length) {
    return "";
  }

  const tagMap = new Map(availableTags.map((tag) => [tag.uuid, tag.name]));
  return tagUUIDs.map((tagId) => tagMap.get(tagId) || tagId).join(", ");
}

function randomCoolColor() {
  const palette = ["#9db8ff", "#8ed7d1", "#c7b2ff", "#9fd7b2", "#9bc5ff"];
  return palette[Math.floor(Math.random() * palette.length)];
}
