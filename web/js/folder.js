function getFolderIdFromQuery() {
  const params = new URLSearchParams(window.location.search);
  return params.get("id");
}

async function loadFolderContents() {
  const folderId = getFolderIdFromQuery();
  const title = document.getElementById("folderTitle");
  const tbody = document.getElementById("folderTableBody");

  if (!folderId) {
    if (title) {
      title.textContent = "Folder belum dipilih";
    }
    if (tbody) {
      tbody.innerHTML = `<tr><td colspan="5">Buka halaman ini dengan query seperti /folder-app?id=123</td></tr>`;
    }
    return;
  }

  if (title) {
    title.textContent = `Folder #${folderId}`;
  }

  tbody.innerHTML = `<tr><td colspan="5">Loading...</td></tr>`;

  try {
    const response = await apiFetch(`/files/folders/${folderId}/contents`);
    const items = response.data || [];

    if (!items.length) {
      tbody.innerHTML = `<tr><td colspan="5">Folder kosong.</td></tr>`;
      return;
    }

    tbody.innerHTML = items
      .map((item) => {
        const itemId = pickField(item, "id", "ID");
        const isFolder = Boolean(pickField(item, "is_folder", "IsFolder"));
        const originalName = pickField(item, "original_name", "OriginalName", "file_name", "FileName") || "Untitled";
        const mimeType = pickField(item, "mime_type", "MimeType") || "File";
        const size = pickField(item, "size", "Size");
        const safeName = escapeHtml(originalName);
        const jsName = escapeJsString(originalName);
        const openAction = isFolder
          ? `window.location.href='/folder-app?id=${itemId}'`
          : `openFileDownload(${itemId}, '${jsName}')`;
        const shareAction = `shareItem(${itemId}, '${jsName}')`;
        const deleteAction = `deleteItem(${itemId}, '${jsName}')`;

        return `
          <tr>
            <td>${safeName}</td>
            <td>${isFolder ? "Folder" : escapeHtml(mimeType)}</td>
            <td>${isFolder ? "-" : formatFileSize(size)}</td>
            <td>${isFolder ? '<span class="badge bg-light-primary">Folder</span>' : '<span class="badge bg-light-success">File</span>'}</td>
            <td>
              <button class="btn btn-sm btn-primary" onclick="${openAction}">Open</button>
              <button class="btn btn-sm btn-outline-primary" onclick="${shareAction}">Share</button>
              <button class="btn btn-sm btn-outline-danger" onclick="${deleteAction}">Delete</button>
            </td>
          </tr>
        `;
      })
      .join("");
  } catch (error) {
    tbody.innerHTML = `<tr><td colspan="5">${escapeHtml(error.message)}</td></tr>`;
  }
}

async function createSubfolder() {
  const folderId = getFolderIdFromQuery();
  if (!folderId) {
    await showError("Folder aktif tidak ditemukan.");
    return;
  }

  const name = await askTextInput("Buat subfolder", "Nama subfolder", "contoh: Foto Keluarga");
  if (!name) {
    return;
  }

  try {
    await apiFetch("/files/folders", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        name,
        parent_id: Number(folderId),
      }),
    });

    await showSuccess(`Subfolder "${name}" berhasil dibuat.`);
    await loadFolderContents();
  } catch (error) {
    await showError(error.message, "Subfolder gagal dibuat");
  }
}

async function uploadFileToFolder() {
  const folderId = getFolderIdFromQuery();
  const input = document.getElementById("folderUploadInput");

  if (!folderId) {
    await showError("Folder aktif tidak ditemukan.");
    return;
  }
  if (!input || !input.files || !input.files[0]) {
    await showError("Pilih file dulu.");
    return;
  }

  const formData = new FormData();
  formData.append("file", input.files[0]);
  formData.append("parent_id", folderId);

  try {
    const response = await authFetch("/files/upload", {
      method: "POST",
      body: formData,
    });
    const data = await response.json();
    if (!response.ok) {
      throw new Error(data.error || data.message || "Upload gagal");
    }

    input.value = "";
    await showSuccess("File berhasil diupload ke folder.");
    await loadFolderContents();
  } catch (error) {
    await showError(error.message, "Upload gagal");
  }
}
