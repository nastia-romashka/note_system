import { useEffect, useRef, useState } from "react";

import { deleteFile, downloadFile, fetchFiles, uploadFile } from "../api/filesApi";
import { PREVIEW_DATA } from "../preview/previewData";

export function useFilesData({ token, workspaceId, selectedNoteId, uiPreview, setMessage, setLoading }) {
  const [files, setFiles] = useState(uiPreview ? PREVIEW_DATA.files["note-1"] : []);
  const [fileDraft, setFileDraft] = useState(null);
  const fileInputRef = useRef(null);

  useEffect(() => {
    if (uiPreview) {
      setFiles(PREVIEW_DATA.files[selectedNoteId] || []);
      return;
    }

    if (!token || !selectedNoteId) {
      setFiles([]);
      return;
    }

    let ignore = false;

    void (async () => {
      try {
        const fileList = await fetchFiles(token, selectedNoteId, workspaceId);
        if (!ignore) {
          setFiles(fileList);
        }
      } catch (error) {
        if (!ignore) {
          handleError(error);
        }
      }
    })();

    return () => {
      ignore = true;
    };
  }, [token, workspaceId, selectedNoteId]);

  async function handleUploadFile(event) {
    event.preventDefault();
    if (uiPreview) {
      setMessage({ type: "info", text: "Preview mode: загрузка файлов отключена." });
      return;
    }

    if (!selectedNoteId || !fileDraft) {
      setMessage({ type: "warning", text: "Выбери заметку и файл для загрузки." });
      return;
    }

    try {
      setLoading(true);
      await uploadFile(token, selectedNoteId, fileDraft, workspaceId);
      setFileDraft(null);
      if (fileInputRef.current) {
        fileInputRef.current.value = "";
      }
      setMessage({ type: "success", text: "Файл прикреплен." });
      setFiles(await fetchFiles(token, selectedNoteId, workspaceId));
    } catch (error) {
      handleError(error);
    } finally {
      setLoading(false);
    }
  }

  async function handleDeleteFile(fileId) {
    if (uiPreview) {
      setMessage({ type: "info", text: "Preview mode: удаление файлов отключено." });
      return;
    }

    if (!selectedNoteId) {
      return;
    }

    try {
      setLoading(true);
      await deleteFile(token, selectedNoteId, fileId, workspaceId);
      setMessage({ type: "success", text: "Файл удален." });
      setFiles(await fetchFiles(token, selectedNoteId, workspaceId));
    } catch (error) {
      handleError(error);
    } finally {
      setLoading(false);
    }
  }

  async function handleDownloadFile(fileId, fileName) {
    if (uiPreview) {
      setMessage({ type: "info", text: `Preview mode: скачивание "${fileName}" недоступно.` });
      return;
    }

    if (!selectedNoteId) {
      return;
    }

    try {
      const blob = await downloadFile(token, selectedNoteId, fileId, workspaceId);
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

  function resetFilesState() {
    setFiles([]);
    setFileDraft(null);
    if (fileInputRef.current) {
      fileInputRef.current.value = "";
    }
  }

  function handleError(error) {
    setMessage({
      type: "error",
      text: error instanceof Error ? error.message : "Произошла ошибка.",
    });
  }

  return {
    files,
    fileDraft,
    setFileDraft,
    fileInputRef,
    handleUploadFile,
    handleDeleteFile,
    handleDownloadFile,
    resetFilesState,
  };
}
