import { useEffect, useMemo, useState } from "react";

import { createCategory, deleteCategory, fetchCategories } from "../api/categoriesApi";
import { createNote, deleteNote, fetchNotes, updateNote } from "../api/notesApi";
import { createTag, fetchTags } from "../api/tagsApi";
import { PREVIEW_DATA } from "../preview/previewData";

export function useNotesData({ token, uiPreview, setMessage, setLoading }) {
  const [categories, setCategories] = useState(uiPreview ? PREVIEW_DATA.categories : []);
  const [notes, setNotes] = useState(uiPreview ? PREVIEW_DATA.notes["cat-1"] : []);
  const [tags, setTags] = useState(uiPreview ? PREVIEW_DATA.tags : []);

  const [selectedCategoryId, setSelectedCategoryId] = useState(uiPreview ? "cat-1" : "");
  const [selectedNoteId, setSelectedNoteId] = useState(uiPreview ? "note-1" : "");
  const [search, setSearch] = useState("");

  const [categoryForm, setCategoryForm] = useState({ name: "", color: "#9db8ff" });
  const [noteForm, setNoteForm] = useState({ header: "", body: "", tags: "" });
  const [noteEditorForm, setNoteEditorForm] = useState({ header: "", body: "", tags: "" });
  const [pendingCategoryDelete, setPendingCategoryDelete] = useState(null);
  const [subcategoryDraft, setSubcategoryDraft] = useState("");
  const [subcategoryParent, setSubcategoryParent] = useState(null);

  const filteredNotes = useMemo(() => {
    const query = search.trim().toLowerCase();
    if (!query) {
      return notes;
    }

    return notes.filter((note) => {
      const haystack = `${note.header} ${note.body} ${note.short_body}`.toLowerCase();
      return haystack.includes(query);
    });
  }, [notes, search]);

  const selectedNote = useMemo(
    () =>
      filteredNotes.find((note) => note.uuid === selectedNoteId) ||
      notes.find((note) => note.uuid === selectedNoteId) ||
      null,
    [filteredNotes, notes, selectedNoteId],
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
  }, [selectedCategoryId]);

  useEffect(() => {
    if (!selectedNote) {
      setNoteEditorForm({ header: "", body: "", tags: "" });
      return;
    }

    setNoteEditorForm({
      header: selectedNote.header || "",
      body: selectedNote.body || "",
      tags: stringifyTagNames(selectedNote.tags || [], tags),
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
      const safeNoteList = Array.isArray(noteList) ? noteList : [];
      setNotes(safeNoteList);

      const nextNoteId = safeNoteList.some((note) => note.uuid === preferredNoteId)
        ? preferredNoteId
        : (safeNoteList[0]?.uuid || "");
      setSelectedNoteId(nextNoteId);
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
      setMessage({ type: "warning", text: "Сначала выбери категорию." });
      return;
    }

    const header = noteForm.header.trim();
    const body = noteForm.body.trim();

    if (!header) {
      setMessage({ type: "warning", text: "Введите заголовок заметки." });
      return;
    }

    if (!body) {
      setMessage({ type: "warning", text: "Введите текст заметки." });
      return;
    }

    try {
      setLoading(true);
      const tagUUIDs = await ensureTagUUIDs(noteForm.tags);
      await createNote(token, {
        header,
        body,
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

    try {
      setLoading(true);
      const tagUUIDs = await ensureTagUUIDs(noteEditorForm.tags);
      await updateNote(token, selectedNote.uuid, {
        header: noteEditorForm.header,
        body: noteEditorForm.body,
        tags: tagUUIDs,
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
    setTags([]);
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
    filteredNotes,
    selectedNote,
    selectedCategoryId,
    setSelectedCategoryId,
    selectedNoteId,
    setSelectedNoteId,
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
    handleCreateCategory,
    openSubcategoryDialog,
    confirmCreateSubcategory,
    openDeleteCategoryDialog,
    confirmDeleteCategory,
    handleCreateNote,
    handleUpdateNote,
    handleDeleteNote,
    resetNotesState,
  };
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
