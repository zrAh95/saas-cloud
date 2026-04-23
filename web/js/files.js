async function loadFiles() {
  const tbody = document.getElementById("filesTableBody");
  if (!tbody) {
    return;
  }

  tbody.innerHTML = `<tr><td colspan="5">Loading...</td></tr>`;

  try {
    const response = await apiFetch("/files/");
    const items = response.data || [];

    if (!items.length) {
      tbody.innerHTML = `<tr><td colspan="5">Belum ada file atau folder.</td></tr>`;
      return;
    }

    tbody.innerHTML = items
      .map((item) => {
        const itemId = pickField(item, "id", "ID");
        const isFolder = Boolean(pickField(item, "is_folder", "IsFolder"));
        const originalName = pickField(item, "original_name", "OriginalName", "file_name", "FileName") || "Untitled";
        const mimeType = pickField(item, "mime_type", "MimeType") || "File";
        const size = pickField(item, "size", "Size");
        const typeLabel = isFolder ? "Folder" : mimeType;
        const sizeLabel = isFolder ? "-" : formatFileSize(size);
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
            <td>${escapeHtml(typeLabel)}</td>
            <td>${sizeLabel}</td>
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

async function createFolder() {
  const name = await askTextInput("Buat folder baru", "Nama folder", "contoh: Dokumen Ayah");
  if (!name) {
    return;
  }

  try {
    await apiFetch("/files/folders", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ name }),
    });

    await showSuccess(`Folder "${name}" berhasil dibuat.`);
    await loadFiles();
  } catch (error) {
    await showError(error.message, "Folder gagal dibuat");
  }
}

async function uploadRootFile() {
  const input = document.getElementById("uploadFileInput");
  if (!input || !input.files || !input.files[0]) {
    await showError("Pilih file dulu.");
    return;
  }

  const formData = new FormData();
  formData.append("file", input.files[0]);

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
    await showSuccess("File berhasil diupload.");
    await loadFiles();
  } catch (error) {
    await showError(error.message, "Upload gagal");
  }
}

async function shareItem(id, name) {
  const email = await askTextInput(`Share "${name}"`, "Masukkan email user tujuan", "user@email.com");
  if (!email) {
    return;
  }

  try {
    await apiFetch(`/files/${id}/share`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ email }),
    });

    await showSuccess(`Akses untuk ${email} berhasil dibuat.`);
  } catch (error) {
    await showError(error.message, "Share gagal");
  }
}

async function deleteItem(id, name) {
  const ok = await askConfirm("Hapus item", `Yakin hapus "${name}"?`, "Ya, hapus");
  if (!ok) {
    return;
  }

  try {
    await apiFetch(`/files/${id}`, {
      method: "DELETE",
    });

    await showSuccess(`"${name}" berhasil dihapus.`);
    await loadFiles();
  } catch (error) {
    await showError(error.message, "Hapus gagal");
  }
}

async function openFileDownload(id, name) {
  try {
    await downloadWithAuth(id, name);
  } catch (error) {
    await showError(error.message, "Buka file gagal");
  }
}
