const API_BASE = "/api/v1";

function getAccessToken() {
  return localStorage.getItem("access_token");
}

function getRefreshToken() {
  return localStorage.getItem("refresh_token");
}

function setTokens(accessToken, refreshToken) {
  localStorage.setItem("access_token", accessToken);
  localStorage.setItem("refresh_token", refreshToken);
}

function clearTokens() {
  localStorage.removeItem("access_token");
  localStorage.removeItem("refresh_token");
}

function pickField(source, ...keys) {
  if (!source || typeof source !== "object") {
    return undefined;
  }

  for (const key of keys) {
    if (source[key] !== undefined && source[key] !== null) {
      return source[key];
    }
  }

  return undefined;
}

async function apiFetch(path, options = {}) {
  const headers = { ...(options.headers || {}) };
  const token = getAccessToken();

  if (token) {
    headers.Authorization = `Bearer ${token}`;
  }

  const response = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers,
  });

  const contentType = response.headers.get("content-type") || "";
  const data = contentType.includes("application/json")
    ? await response.json()
    : await response.text();

  if (!response.ok) {
    if (response.status === 401) {
      clearTokens();
      if (window.location.pathname !== "/login") {
        window.location.href = "/login";
      }
    }

    throw new Error(data.error || data.message || "Request gagal");
  }

  return data;
}

async function authFetch(path, options = {}) {
  const headers = { ...(options.headers || {}) };
  const token = getAccessToken();

  if (token) {
    headers.Authorization = `Bearer ${token}`;
  }

  return fetch(`${API_BASE}${path}`, {
    ...options,
    headers,
  });
}

async function showAlert(icon, title, text) {
  if (typeof Swal !== "undefined") {
    await Swal.fire({
      icon,
      title,
      text,
      confirmButtonText: "OK",
      confirmButtonColor: "#435ebe",
      background: "#1e1e2d",
      color: "#f2f7ff",
    });
    return;
  }

  alert(text || title);
}

async function showError(message, title = "Terjadi masalah") {
  await showAlert("error", title, message);
}

async function showSuccess(message, title = "Berhasil") {
  await showAlert("success", title, message);
}

async function askTextInput(title, inputLabel, placeholder = "") {
  if (typeof Swal !== "undefined") {
    const result = await Swal.fire({
      title,
      input: "text",
      inputLabel,
      inputPlaceholder: placeholder,
      showCancelButton: true,
      confirmButtonText: "Simpan",
      cancelButtonText: "Batal",
      confirmButtonColor: "#435ebe",
      background: "#1e1e2d",
      color: "#f2f7ff",
      inputValidator: (value) => {
        if (!String(value || "").trim()) {
          return "Wajib diisi";
        }
        return null;
      },
    });

    return result.isConfirmed ? String(result.value || "").trim() : null;
  }

  const value = prompt(inputLabel);
  return value ? value.trim() : null;
}

async function askConfirm(title, text, confirmButtonText = "Lanjutkan") {
  if (typeof Swal !== "undefined") {
    const result = await Swal.fire({
      title,
      text,
      icon: "warning",
      showCancelButton: true,
      confirmButtonText,
      cancelButtonText: "Batal",
      confirmButtonColor: "#dc3545",
      background: "#1e1e2d",
      color: "#f2f7ff",
    });

    return result.isConfirmed;
  }

  return confirm(text);
}

async function downloadWithAuth(fileId, fallbackName = "download") {
  const response = await authFetch(`/files/${fileId}/download`, {
    method: "GET",
  });

  const contentType = response.headers.get("content-type") || "";
  if (!response.ok) {
    const payload = contentType.includes("application/json")
      ? await response.json()
      : { message: await response.text() };
    throw new Error(payload.error || payload.message || "Download gagal");
  }

  const blob = await response.blob();
  const disposition = response.headers.get("content-disposition") || "";
  const matchedName = disposition.match(/filename="?([^"]+)"?/i);
  const fileName = matchedName?.[1] || fallbackName;
  const blobUrl = window.URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = blobUrl;
  link.download = fileName;
  document.body.appendChild(link);
  link.click();
  link.remove();
  window.URL.revokeObjectURL(blobUrl);
}

function escapeHtml(value) {
  return String(value ?? "")
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#039;");
}

function escapeJsString(value) {
  return String(value ?? "")
    .replace(/\\/g, "\\\\")
    .replace(/'/g, "\\'")
    .replace(/"/g, '\\"')
    .replace(/\r/g, "\\r")
    .replace(/\n/g, "\\n");
}

function formatFileSize(size) {
  const bytes = Number(size || 0);
  if (bytes < 1024) {
    return `${bytes} B`;
  }
  if (bytes < 1024 * 1024) {
    return `${(bytes / 1024).toFixed(1)} KB`;
  }
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

function formatGigabytes(bytes) {
  const gb = Number(bytes || 0) / (1024 * 1024 * 1024);
  return `${gb.toFixed(2)} GB`;
}

function setText(id, value) {
  const element = document.getElementById(id);
  if (element) {
    element.textContent = value;
  }
}
