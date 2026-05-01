import { useEffect, useMemo, useState } from "react";

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || "";
const UI_PREVIEW = import.meta.env.VITE_UI_PREVIEW === "true";
const TOKEN_KEY = "note-system-token";
const REFRESH_TOKEN_KEY = "note-system-refresh-token";

const PREVIEW_DATA = {
  categories: [
    { uuid: "cat-1", name: "Категория 1", color: "#9db8ff" },
    { uuid: "cat-1-1", name: "1.1 Подкатегория", color: "#8ed7d1" },
    { uuid: "cat-1-1-1", name: "1.1.1 Подкатегория", color: "#c7b2ff" },
    { uuid: "cat-2", name: "Категория 2", color: "#9fd7b2" },
  ],
  tags: [
    { uuid: "tag-1", name: "tag1" },
    { uuid: "tag-2", name: "tag2" },
    { uuid: "tag-3", name: "tag3" },
  ],
  notes: {
    "cat-1": [
      {
        uuid: "note-1",
        header: "Имя заметки",
        body:
          "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.",
        short_body:
          "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.",
        tags: ["tag-1", "tag-2", "tag-3"],
      },
    ],
    "cat-1-1": [
      {
        uuid: "note-2",
        header: "Заметка по подкатегории",
        body:
          "Структурируй мысли по вложенным категориям и собирай связанные материалы в одном месте. Такой режим помогает посмотреть будущий интерфейс без запуска backend.",
        short_body:
          "Структурируй мысли по вложенным категориям и собирай связанные материалы в одном месте.",
        tags: ["tag-2"],
      },
    ],
    "cat-1-1-1": [],
    "cat-2": [
      {
        uuid: "note-3",
        header: "Вторая категория",
        body:
          "Отдельная область для рабочих заметок, быстрых черновиков и небольших материалов, которые нужны в течение дня.",
        short_body:
          "Отдельная область для рабочих заметок, быстрых черновиков и небольших материалов.",
        tags: [],
      },
    ],
  },
  files: {
    "note-1": [
      { id: "file-1", name: "image.png", size: 183421, content_type: "image/png" },
      { id: "file-2", name: "brief.pdf", size: 94213, content_type: "application/pdf" },
      { id: "file-3", name: "Link.txt", size: 1024, content_type: "text/plain" },
    ],
    "note-2": [{ id: "file-4", name: "roadmap.docx", size: 42018, content_type: "application/vnd.openxmlformats-officedocument.wordprocessingml.document" }],
    "note-3": [],
  },
};

function App() {
  const storedToken = localStorage.getItem(TOKEN_KEY) || sessionStorage.getItem(TOKEN_KEY) || "";
  const storedRefresh = localStorage.getItem(REFRESH_TOKEN_KEY) || sessionStorage.getItem(REFRESH_TOKEN_KEY) || "";

  const [page, setPage] = useState(() => (UI_PREVIEW ? "login" : storedToken ? "notes" : "login"));
  const [token, setToken] = useState(() => (UI_PREVIEW ? "preview-token" : storedToken));
  const [refreshToken, setRefreshToken] = useState(() => (UI_PREVIEW ? "preview-refresh-token" : storedRefresh));
  const [message, setMessage] = useState(null);
  const [loading, setLoading] = useState(false);

  const [categories, setCategories] = useState(UI_PREVIEW ? PREVIEW_DATA.categories : []);
  const [notes, setNotes] = useState(UI_PREVIEW ? PREVIEW_DATA.notes["cat-1"] : []);
  const [tags, setTags] = useState(UI_PREVIEW ? PREVIEW_DATA.tags : []);
  const [files, setFiles] = useState(UI_PREVIEW ? PREVIEW_DATA.files["note-1"] : []);

  const [selectedCategoryId, setSelectedCategoryId] = useState(UI_PREVIEW ? "cat-1" : "");
  const [selectedNoteId, setSelectedNoteId] = useState(UI_PREVIEW ? "note-1" : "");
  const [search, setSearch] = useState("");

  const [loginForm, setLoginForm] = useState({ username: "", password: "", remember: true });
  const [signupForm, setSignupForm] = useState({ username: "", email: "", password: "" });
  const [categoryForm, setCategoryForm] = useState({ name: "", color: "#9db8ff" });
  const [tagForm, setTagForm] = useState({ name: "" });
  const [noteForm, setNoteForm] = useState({ header: "", body: "", tags: "" });
  const [fileDraft, setFileDraft] = useState(null);

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
    () => filteredNotes.find((note) => note.uuid === selectedNoteId) || notes.find((note) => note.uuid === selectedNoteId) || null,
    [filteredNotes, notes, selectedNoteId],
  );

  useEffect(() => {
    if (UI_PREVIEW || !token) {
      return;
    }

    void bootstrap();
  }, [token]);

  useEffect(() => {
    if (UI_PREVIEW) {
      const nextNotes = PREVIEW_DATA.notes[selectedCategoryId] || [];
      setNotes(nextNotes);
      const nextNoteId = nextNotes.find((note) => note.uuid === selectedNoteId)?.uuid || nextNotes[0]?.uuid || "";
      setSelectedNoteId(nextNoteId);
      setFiles(PREVIEW_DATA.files[nextNoteId] || []);
      return;
    }

    if (!token || !selectedCategoryId) {
      setNotes([]);
      setSelectedNoteId("");
      setFiles([]);
      return;
    }

    void loadNotes(selectedCategoryId, selectedNoteId);
  }, [selectedCategoryId]);

  useEffect(() => {
    if (UI_PREVIEW) {
      setFiles(PREVIEW_DATA.files[selectedNoteId] || []);
      return;
    }

    if (!token || !selectedNoteId) {
      setFiles([]);
      return;
    }

    void (async () => {
      try {
        const fileList = await fetchFiles(selectedNoteId);
        setFiles(fileList);
      } catch (error) {
        handleError(error);
      }
    })();
  }, [selectedNoteId]);

  async function bootstrap() {
    try {
      setLoading(true);
      setMessage(null);

      const [categoryList, tagList] = await Promise.all([fetchCategories(), fetchTags()]);
      setCategories(categoryList);
      setTags(tagList);

      const nextCategory = categoryList.find((item) => item.uuid === selectedCategoryId)?.uuid || categoryList[0]?.uuid || "";
      setSelectedCategoryId(nextCategory);

      if (nextCategory) {
        await loadNotes(nextCategory, selectedNoteId);
      } else {
        setNotes([]);
        setSelectedNoteId("");
        setFiles([]);
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
      const noteList = await fetchNotes(categoryId);
      setNotes(noteList);

      const nextNoteId = noteList.some((note) => note.uuid === preferredNoteId)
        ? preferredNoteId
        : (noteList[0]?.uuid || "");
      setSelectedNoteId(nextNoteId);

      if (nextNoteId) {
        const fileList = await fetchFiles(nextNoteId);
        setFiles(fileList);
      } else {
        setFiles([]);
      }
    } catch (error) {
      handleError(error);
    } finally {
      setLoading(false);
    }
  }

  function persistSession(nextToken, nextRefreshToken, remember = true) {
    setToken(nextToken);
    setRefreshToken(nextRefreshToken);
    setPage("notes");

    if (UI_PREVIEW) {
      return;
    }

    if (remember) {
      localStorage.setItem(TOKEN_KEY, nextToken);
      localStorage.setItem(REFRESH_TOKEN_KEY, nextRefreshToken);
      sessionStorage.removeItem(TOKEN_KEY);
      sessionStorage.removeItem(REFRESH_TOKEN_KEY);
    } else {
      sessionStorage.setItem(TOKEN_KEY, nextToken);
      sessionStorage.setItem(REFRESH_TOKEN_KEY, nextRefreshToken);
      localStorage.removeItem(TOKEN_KEY);
      localStorage.removeItem(REFRESH_TOKEN_KEY);
    }
  }

  function logout() {
    if (UI_PREVIEW) {
      setPage("login");
      setMessage({ type: "info", text: "Режим предпросмотра: возврат на экран входа." });
      return;
    }

    setToken("");
    setRefreshToken("");
    setCategories([]);
    setNotes([]);
    setTags([]);
    setFiles([]);
    setSelectedCategoryId("");
    setSelectedNoteId("");
    setSearch("");
    setPage("login");
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(REFRESH_TOKEN_KEY);
    sessionStorage.removeItem(TOKEN_KEY);
    sessionStorage.removeItem(REFRESH_TOKEN_KEY);
    setMessage({ type: "info", text: "Сессия завершена." });
  }

  async function handleLogin(event) {
    event.preventDefault();
    if (UI_PREVIEW) {
      persistSession("preview-token", "preview-refresh-token", loginForm.remember);
      setMessage({ type: "success", text: "Preview mode: открыт экран заметок." });
      return;
    }

    try {
      setLoading(true);
      const authData = await request("/api/auth", {
        method: "POST",
        body: JSON.stringify({
          username: loginForm.username,
          password: loginForm.password,
        }),
      });
      persistSession(authData.token, authData.refresh_token, loginForm.remember);
      setMessage({ type: "success", text: "Вход выполнен." });
    } catch (error) {
      handleError(error);
    } finally {
      setLoading(false);
    }
  }

  async function handleSignup(event) {
    event.preventDefault();
    if (UI_PREVIEW) {
      persistSession("preview-token", "preview-refresh-token", true);
      setMessage({ type: "success", text: "Preview mode: открыт экран заметок после регистрации." });
      return;
    }

    try {
      setLoading(true);
      const authData = await request("/api/signup", {
        method: "POST",
        body: JSON.stringify(signupForm),
      });
      persistSession(authData.token, authData.refresh_token, true);
      setMessage({ type: "success", text: "Аккаунт создан." });
    } catch (error) {
      handleError(error);
    } finally {
      setLoading(false);
    }
  }

  async function handleCreateCategory(event) {
    event.preventDefault();
    if (UI_PREVIEW) {
      setMessage({ type: "info", text: "Preview mode: создание категории отключено." });
      return;
    }

    try {
      setLoading(true);
      await request("/api/categories", {
        method: "POST",
        headers: authHeaders(token),
        body: JSON.stringify(categoryForm),
      });
      setCategoryForm({ name: "", color: randomCoolColor() });
      setMessage({ type: "success", text: "Категория создана." });
      await bootstrap();
    } catch (error) {
      handleError(error);
    } finally {
      setLoading(false);
    }
  }

  async function handleCreateTag(event) {
    event.preventDefault();
    if (UI_PREVIEW) {
      setMessage({ type: "info", text: "Preview mode: добавление тегов отключено." });
      return;
    }

    try {
      setLoading(true);
      await request("/api/tags", {
        method: "POST",
        headers: authHeaders(token),
        body: JSON.stringify(tagForm),
      });
      setTagForm({ name: "" });
      setMessage({ type: "success", text: "Тег добавлен." });
      const nextTags = await fetchTags();
      setTags(nextTags);
    } catch (error) {
      handleError(error);
    } finally {
      setLoading(false);
    }
  }

  async function handleCreateNote(event) {
    event.preventDefault();
    if (UI_PREVIEW) {
      setMessage({ type: "info", text: "Preview mode: создание заметок отключено." });
      return;
    }

    if (!selectedCategoryId) {
      setMessage({ type: "warning", text: "Сначала выбери категорию." });
      return;
    }

    try {
      setLoading(true);
      await request("/api/notes", {
        method: "POST",
        headers: authHeaders(token),
        body: JSON.stringify({
          header: noteForm.header,
          body: noteForm.body,
          category_uuid: selectedCategoryId,
          tags: parseTags(noteForm.tags, tags),
        }),
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

  async function handleDeleteNote(noteId) {
    if (UI_PREVIEW) {
      setMessage({ type: "info", text: "Preview mode: удаление заметок отключено." });
      return;
    }

    try {
      setLoading(true);
      await request(`/api/notes/${noteId}`, {
        method: "DELETE",
        headers: authHeaders(token),
      });
      setMessage({ type: "success", text: "Заметка удалена." });
      await loadNotes(selectedCategoryId, "");
    } catch (error) {
      handleError(error);
    } finally {
      setLoading(false);
    }
  }

  async function handleUploadFile(event) {
    event.preventDefault();
    if (UI_PREVIEW) {
      setMessage({ type: "info", text: "Preview mode: загрузка файлов отключена." });
      return;
    }

    if (!selectedNoteId || !fileDraft) {
      setMessage({ type: "warning", text: "Выбери заметку и файл для загрузки." });
      return;
    }

    try {
      setLoading(true);
      const formData = new FormData();
      formData.append("file", fileDraft);
      await request(`/api/notes/${selectedNoteId}/files`, {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
        },
        body: formData,
      });
      setFileDraft(null);
      setMessage({ type: "success", text: "Файл прикреплен." });
      const nextFiles = await fetchFiles(selectedNoteId);
      setFiles(nextFiles);
    } catch (error) {
      handleError(error);
    } finally {
      setLoading(false);
    }
  }

  async function handleDeleteFile(fileId) {
    if (UI_PREVIEW) {
      setMessage({ type: "info", text: "Preview mode: удаление файлов отключено." });
      return;
    }

    if (!selectedNoteId) {
      return;
    }

    try {
      setLoading(true);
      await request(`/api/notes/${selectedNoteId}/files/${fileId}`, {
        method: "DELETE",
        headers: authHeaders(token),
      });
      setMessage({ type: "success", text: "Файл удален." });
      const nextFiles = await fetchFiles(selectedNoteId);
      setFiles(nextFiles);
    } catch (error) {
      handleError(error);
    } finally {
      setLoading(false);
    }
  }

  async function handleDownloadFile(fileId, fileName) {
    if (UI_PREVIEW) {
      setMessage({ type: "info", text: `Preview mode: скачивание "${fileName}" недоступно.` });
      return;
    }

    if (!selectedNoteId) {
      return;
    }

    try {
      const response = await fetch(`${API_BASE_URL}/api/notes/${selectedNoteId}/files/${fileId}`, {
        headers: authHeaders(token, "*/*"),
      });

      if (!response.ok) {
        const errorData = await safeJson(response);
        throw new Error(errorData?.message || "Не удалось скачать файл.");
      }

      const blob = await response.blob();
      const url = URL.createObjectURL(blob);
      const link = document.createElement("a");
      link.href = url;
      link.download = fileName || "attachment";
      document.body.appendChild(link);
      link.click();
      link.remove();
      URL.revokeObjectURL(url);
    } catch (error) {
      handleError(error);
    }
  }

  async function fetchCategories() {
    return request("/api/categories", { headers: authHeaders(token) });
  }

  async function fetchTags() {
    return request("/api/tags", { headers: authHeaders(token) });
  }

  async function fetchNotes(categoryId) {
    return request(`/api/notes?category_uuid=${encodeURIComponent(categoryId)}`, {
      headers: authHeaders(token),
    }).catch((error) => {
      if (String(error.message).toLowerCase().includes("not found")) {
        return [];
      }
      throw error;
    });
  }

  async function fetchFiles(noteId) {
    return request(`/api/notes/${noteId}/files`, { headers: authHeaders(token) });
  }

  function handleError(error) {
    setMessage({
      type: "error",
      text: error instanceof Error ? error.message : "Произошла ошибка.",
    });
  }

  return (
    <div className="site-shell">
      <div className="page-glow page-glow-left" />
      <div className="page-glow page-glow-right" />

      {UI_PREVIEW && (
        <div className="preview-banner">
          <span>Preview mode</span>
          <div className="preview-switch">
            <button type="button" className={page === "login" ? "active" : ""} onClick={() => setPage("login")}>
              Вход
            </button>
            <button type="button" className={page === "signup" ? "active" : ""} onClick={() => setPage("signup")}>
              Регистрация
            </button>
            <button type="button" className={page === "notes" ? "active" : ""} onClick={() => setPage("notes")}>
              Заметки
            </button>
          </div>
        </div>
      )}

      {message && <FlashMessage message={message} onClose={() => setMessage(null)} />}

      {page === "login" && (
        <AuthPage
          title="Вход"
          actionLabel="Войти"
          footerLabel="Нет аккаунта?"
          footerAction="Регистрация"
          onFooterClick={() => setPage("signup")}
          onSubmit={handleLogin}
          loading={loading}
        >
          <LabeledInput
            label="Email / Username"
            value={loginForm.username}
            placeholder="Введите логин"
            onChange={(value) => setLoginForm((current) => ({ ...current, username: value }))}
          />
          <LabeledInput
            label="Пароль"
            type="password"
            value={loginForm.password}
            placeholder="Password"
            onChange={(value) => setLoginForm((current) => ({ ...current, password: value }))}
          />
          <label className="remember-row">
            <input
              type="checkbox"
              checked={loginForm.remember}
              onChange={(event) => setLoginForm((current) => ({ ...current, remember: event.target.checked }))}
            />
            <span>Remember me</span>
          </label>
        </AuthPage>
      )}

      {page === "signup" && (
        <AuthPage
          title="Регистрация"
          actionLabel="Создать аккаунт"
          footerLabel="Уже есть аккаунт?"
          footerAction="Вход"
          onFooterClick={() => setPage("login")}
          onSubmit={handleSignup}
          loading={loading}
        >
          <LabeledInput
            label="Username"
            value={signupForm.username}
            placeholder="Придумай логин"
            onChange={(value) => setSignupForm((current) => ({ ...current, username: value }))}
          />
          <LabeledInput
            label="Email"
            type="email"
            value={signupForm.email}
            placeholder="mail@example.com"
            onChange={(value) => setSignupForm((current) => ({ ...current, email: value }))}
          />
          <LabeledInput
            label="Пароль"
            type="password"
            value={signupForm.password}
            placeholder="Не менее 6 символов"
            onChange={(value) => setSignupForm((current) => ({ ...current, password: value }))}
          />
        </AuthPage>
      )}

      {page === "notes" && (
        <NotesPage
          loading={loading}
          categories={categories}
          selectedCategoryId={selectedCategoryId}
          onSelectCategory={setSelectedCategoryId}
          notes={filteredNotes}
          selectedNote={selectedNote}
          selectedNoteId={selectedNoteId}
          onSelectNote={setSelectedNoteId}
          onLogout={logout}
          search={search}
          onSearch={setSearch}
          categoryForm={categoryForm}
          onCategoryFormChange={setCategoryForm}
          onCreateCategory={handleCreateCategory}
          noteForm={noteForm}
          onNoteFormChange={setNoteForm}
          onCreateNote={handleCreateNote}
          onDeleteNote={handleDeleteNote}
          tagForm={tagForm}
          onTagFormChange={setTagForm}
          onCreateTag={handleCreateTag}
          files={files}
          onDownloadFile={handleDownloadFile}
          onDeleteFile={handleDeleteFile}
          onUploadFile={handleUploadFile}
          onFilePick={setFileDraft}
          fileDraft={fileDraft}
        />
      )}
    </div>
  );
}

function AuthPage({
  title,
  actionLabel,
  footerLabel,
  footerAction,
  onFooterClick,
  onSubmit,
  loading,
  children,
}) {
  return (
    <main className="auth-page">
      <section className="auth-panel">
        <header className="auth-header">
          <h1>{title}</h1>
        </header>
        <form className="auth-form" onSubmit={onSubmit}>
          {children}
          <button className="primary-button" type="submit" disabled={loading}>
            {actionLabel}
          </button>
        </form>
        <div className="auth-footer">
          <span>{footerLabel}</span>
          <button className="text-button" onClick={onFooterClick} type="button">
            {footerAction}
          </button>
        </div>
      </section>
    </main>
  );
}

function NotesPage({
  loading,
  categories,
  selectedCategoryId,
  onSelectCategory,
  notes,
  selectedNote,
  selectedNoteId,
  onSelectNote,
  onLogout,
  search,
  onSearch,
  categoryForm,
  onCategoryFormChange,
  onCreateCategory,
  noteForm,
  onNoteFormChange,
  onCreateNote,
  onDeleteNote,
  tagForm,
  onTagFormChange,
  onCreateTag,
  files,
  onDownloadFile,
  onDeleteFile,
  onUploadFile,
  onFilePick,
  fileDraft,
}) {
  return (
    <main className="notes-page">
      <header className="notes-header">
        <h1>Заметки</h1>
        <div className="notes-toolbar">
          <label className="search-box">
            <span>⌕</span>
            <input value={search} onChange={(event) => onSearch(event.target.value)} placeholder="Search..." />
          </label>
          <button className="secondary-button" onClick={onLogout} type="button">
            Выйти
          </button>
        </div>
      </header>

      <section className="notes-layout">
        <aside className="sidebar-panel">
          <div className="panel-title">Категории</div>
          <form className="compact-form" onSubmit={onCreateCategory}>
            <input
              value={categoryForm.name}
              onChange={(event) =>
                onCategoryFormChange((current) => ({ ...current, name: event.target.value }))
              }
              placeholder="Новая категория"
            />
            <div className="compact-row">
              <input
                type="color"
                value={categoryForm.color}
                onChange={(event) =>
                  onCategoryFormChange((current) => ({ ...current, color: event.target.value }))
                }
              />
              <button className="secondary-button" type="submit">
                Добавить
              </button>
            </div>
          </form>
          <div className="category-list">
            {categories.map((category, index) => (
              <button
                key={category.uuid}
                className={`category-item ${selectedCategoryId === category.uuid ? "active" : ""}`}
                onClick={() => onSelectCategory(category.uuid)}
              >
                <span className="category-index">{index + 1}.</span>
                <span>{category.name}</span>
              </button>
            ))}
            {!categories.length && <div className="empty-copy">Категории пока не созданы.</div>}
          </div>
        </aside>

        <section className="content-panel">
          <div className="content-split">
            <div className="editor-column">
              <form className="note-creator" onSubmit={onCreateNote}>
                <input
                  value={noteForm.header}
                  onChange={(event) =>
                    onNoteFormChange((current) => ({ ...current, header: event.target.value }))
                  }
                  placeholder="Имя заметки"
                />
                <textarea
                  rows={6}
                  value={noteForm.body}
                  onChange={(event) =>
                    onNoteFormChange((current) => ({ ...current, body: event.target.value }))
                  }
                  placeholder="Текст заметки"
                />
                <input
                  value={noteForm.tags}
                  onChange={(event) =>
                    onNoteFormChange((current) => ({ ...current, tags: event.target.value }))
                  }
                  placeholder="Теги через запятую"
                />
                <button className="primary-button" type="submit">
                  Создать заметку
                </button>
              </form>
              <div className="notes-list">
                {notes.map((note) => (
                  <article
                    key={note.uuid}
                    className={`note-preview ${selectedNoteId === note.uuid ? "active" : ""}`}
                    onClick={() => onSelectNote(note.uuid)}
                  >
                    <div className="note-preview-top">
                      <h2>{note.header}</h2>
                      <button
                        className="text-button danger"
                        type="button"
                        onClick={(event) => {
                          event.stopPropagation();
                          void onDeleteNote(note.uuid);
                        }}
                      >
                        удалить
                      </button>
                    </div>
                    <div className="note-preview-line" />
                    <p>{note.short_body || note.body}</p>
                  </article>
                ))}
                {!notes.length && <div className="empty-copy">Для выбранной категории заметок пока нет.</div>}
              </div>
            </div>
            <aside className="detail-column">
              {selectedNote ? (
                <div className="note-card-sheet">
                  <header className="sheet-header">
                    <h2>{selectedNote.header}</h2>
                    <span>{loading ? "Сохранение..." : new Date().toLocaleDateString("ru-RU")}</span>
                  </header>
                  <div className="sheet-body">{selectedNote.body}</div>
                  <div className="tag-section">
                    <div className="side-title">@tags</div>
                    <div className="tag-list">
                      {(selectedNote.tags || []).map((tagId) => (
                        <span key={tagId}>@{tagId.slice(0, 8)}</span>
                      ))}
                      {!selectedNote.tags?.length && <span>@empty</span>}
                    </div>
                    <form className="compact-form" onSubmit={onCreateTag}>
                      <input
                        value={tagForm.name}
                        onChange={(event) => onTagFormChange({ name: event.target.value })}
                        placeholder="Новый тег"
                      />
                      <button className="secondary-button" type="submit">
                        Добавить тег
                      </button>
                    </form>
                  </div>
                  <div className="attachments">
                    <div className="side-title">Вложения</div>
                    <form className="attachment-form" onSubmit={onUploadFile}>
                      <input type="file" onChange={(event) => onFilePick(event.target.files?.[0] || null)} />
                      <button className="secondary-button" type="submit">
                        {fileDraft ? "Прикрепить" : "Загрузить"}
                      </button>
                    </form>
                    <div className="attachment-list">
                      {files.map((file) => (
                        <div className="attachment-item" key={file.id}>
                          <div>
                            <strong>{file.name}</strong>
                            <span>{formatSize(file.size)}</span>
                          </div>
                          <div className="attachment-actions">
                            <button className="text-button" type="button" onClick={() => onDownloadFile(file.id, file.name)}>
                              Link
                            </button>
                            <button className="text-button danger" type="button" onClick={() => onDeleteFile(file.id)}>
                              удалить
                            </button>
                          </div>
                        </div>
                      ))}
                      {!files.length && <div className="empty-copy">Вложений пока нет.</div>}
                    </div>
                  </div>
                </div>
              ) : (
                <div className="note-card-sheet empty-sheet">
                  <div className="empty-copy">Выбери заметку, чтобы увидеть содержимое и вложения.</div>
                </div>
              )}
            </aside>
          </div>
        </section>
      </section>
    </main>
  );
}

function LabeledInput({ label, value, onChange, placeholder, type = "text" }) {
  return (
    <label className="field-row">
      <span>{label}</span>
      <input type={type} value={value} onChange={(event) => onChange(event.target.value)} placeholder={placeholder} />
    </label>
  );
}

function FlashMessage({ message, onClose }) {
  return (
    <div className={`flash-message ${message.type}`}>
      <span>{message.text}</span>
      <button type="button" onClick={onClose}>
        закрыть
      </button>
    </div>
  );
}

function authHeaders(token, accept = "application/json") {
  return {
    Accept: accept,
    Authorization: `Bearer ${token}`,
  };
}

async function request(path, options = {}) {
  const response = await fetch(`${API_BASE_URL}${path}`, options);
  if (!response.ok) {
    const errorData = await safeJson(response);
    throw new Error(errorData?.developer_message || errorData?.message || "Запрос завершился с ошибкой.");
  }
  if (response.status === 204) {
    return null;
  }
  const contentType = response.headers.get("Content-Type") || "";
  if (contentType.includes("application/json")) {
    return response.json();
  }
  return response.text();
}

async function safeJson(response) {
  try {
    return await response.json();
  } catch {
    return null;
  }
}

function parseTags(rawTags, availableTags) {
  if (!rawTags.trim()) {
    return [];
  }
  const names = rawTags
    .split(",")
    .map((item) => item.trim().toLowerCase())
    .filter(Boolean);
  return availableTags.filter((tag) => names.includes(tag.name.toLowerCase())).map((tag) => tag.uuid);
}

function formatSize(size) {
  if (!size) {
    return "0 B";
  }
  if (size < 1024) {
    return `${size} B`;
  }
  if (size < 1024 * 1024) {
    return `${(size / 1024).toFixed(1)} KB`;
  }
  return `${(size / (1024 * 1024)).toFixed(1)} MB`;
}

function randomCoolColor() {
  const palette = ["#9db8ff", "#8ed7d1", "#c7b2ff", "#9fd7b2", "#9bc5ff"];
  return palette[Math.floor(Math.random() * palette.length)];
}

export default App;
