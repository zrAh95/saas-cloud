async function loadSharedFiles() {
  const tbody = document.getElementById("sharedTableBody");
  if (!tbody) {
    return;
  }

  tbody.innerHTML = `<tr><td colspan="5">Loading...</td></tr>`;

  try {
    const response = await apiFetch("/files/shared");
    const items = response.data || [];

    if (!items.length) {
      tbody.innerHTML = `<tr><td colspan="5">Belum ada incoming share.</td></tr>`;
      return;
    }

    tbody.innerHTML = items
      .map((item) => {
        const shareId = pickField(item, "share_id", "ShareID", "shareId");
        const fileId = pickField(item, "file_id", "FileID", "fileId");
        const isFolder = Boolean(pickField(item, "is_folder", "IsFolder"));
        const status = pickField(item, "status", "Status") || "-";
        const originalName = pickField(item, "original_name", "OriginalName", "file_name", "FileName") || "Untitled";
        const ownerEmail = pickField(item, "owner_email", "OwnerEmail") || "-";
        const jsName = escapeJsString(originalName);
        const acceptButton =
          status === "pending"
            ? `<button class="btn btn-sm btn-primary" onclick="acceptShare(${shareId})">Accept</button>`
            : `<span class="badge bg-light-success">Accepted</span>`;
        const openButton =
          status === "accepted"
            ? `<button class="btn btn-sm btn-outline-primary" onclick="openSharedItem(${fileId}, ${isFolder}, '${jsName}')">Open</button>`
            : "";

        return `
          <tr>
            <td>${escapeHtml(originalName)}</td>
            <td>${escapeHtml(ownerEmail)}</td>
            <td>${isFolder ? "Folder" : "File"}</td>
            <td>${escapeHtml(status)}</td>
            <td>${acceptButton} ${openButton}</td>
          </tr>
        `;
      })
      .join("");
  } catch (error) {
    tbody.innerHTML = `<tr><td colspan="5">${escapeHtml(error.message)}</td></tr>`;
  }
}

async function acceptShare(shareId) {
  try {
    await apiFetch(`/files/shares/${shareId}/accept`, {
      method: "POST",
    });

    await showSuccess("Share berhasil diterima.");
    await loadSharedFiles();
  } catch (error) {
    await showError(error.message, "Accept share gagal");
  }
}

async function openSharedItem(fileId, isFolder, name) {
  if (isFolder) {
    window.location.href = `/folder-app?id=${fileId}`;
    return;
  }

  try {
    await downloadWithAuth(fileId, name || "shared-file");
  } catch (error) {
    await showError(error.message, "Buka file gagal");
  }
}
