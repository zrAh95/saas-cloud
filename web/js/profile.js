async function loadProfile() {
  try {
    const response = await apiFetch("/auth/profile");
    const profile = response.data || {};

    setInputValue("profileName", pickField(profile, "name") || "");
    setInputValue("profileEmail", pickField(profile, "email") || "");
    setInputValue("profilePhone", pickField(profile, "phone") || "");
    setText("profileUserId", `#${pickField(profile, "id") || "-"}`);
    setText("profileVerified", Number(pickField(profile, "is_verified") || 0) === 1 ? "Verified" : "Not Verified");
    setText("profileCreatedAt", formatDateTime(pickField(profile, "created_at")));
    await loadProfileAvatar(Boolean(pickField(profile, "has_avatar")));
  } catch (error) {
    await showError(error.message, "Profile gagal dimuat");
  }
}

async function updateProfile(event) {
  event.preventDefault();

  const payload = {
    name: document.getElementById("profileName").value.trim(),
    email: document.getElementById("profileEmail").value.trim(),
    phone: document.getElementById("profilePhone").value.trim(),
    current_password: document.getElementById("currentPassword").value,
    new_password: document.getElementById("newPassword").value,
  };

  try {
    const response = await apiFetch("/auth/profile", {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(payload),
    });

    document.getElementById("currentPassword").value = "";
    document.getElementById("newPassword").value = "";
    await showSuccess(response.message || "Profile berhasil diupdate.");
    await loadProfile();
  } catch (error) {
    await showError(error.message, "Update profile gagal");
  }
}

async function uploadProfileAvatar() {
  const input = document.getElementById("profileAvatarInput");
  if (!input || !input.files || !input.files[0]) {
    await showError("Pilih foto dulu.");
    return;
  }

  const formData = new FormData();
  formData.append("avatar", input.files[0]);

  try {
    const response = await authFetch("/auth/profile/avatar", {
      method: "POST",
      body: formData,
    });
    const data = await response.json();
    if (!response.ok) {
      throw new Error(data.error || data.message || "Upload foto profile gagal");
    }

    input.value = "";
    await showSuccess(data.message || "Foto profile berhasil diupdate.");
    await loadProfileAvatar(true);
  } catch (error) {
    await showError(error.message, "Upload foto gagal");
  }
}

async function loadProfileAvatar(hasAvatar) {
  const image = document.getElementById("profileAvatarImage");
  const fallback = document.getElementById("profileAvatarFallback");

  if (!image || !fallback) {
    return;
  }

  if (!hasAvatar) {
    image.classList.add("d-none");
    fallback.classList.remove("d-none");
    return;
  }

  try {
    const response = await authFetch(`/auth/profile/avatar?t=${Date.now()}`);
    if (!response.ok) {
      throw new Error("Avatar belum tersedia");
    }

    const blob = await response.blob();
    image.src = URL.createObjectURL(blob);
    image.classList.remove("d-none");
    fallback.classList.add("d-none");
  } catch (error) {
    image.classList.add("d-none");
    fallback.classList.remove("d-none");
  }
}

function setInputValue(id, value) {
  const element = document.getElementById(id);
  if (element) {
    element.value = value;
  }
}

function setText(id, value) {
  const element = document.getElementById(id);
  if (element) {
    element.textContent = value;
  }
}

function formatDateTime(value) {
  if (!value) {
    return "-";
  }

  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return "-";
  }

  return date.toLocaleString("id-ID", {
    dateStyle: "medium",
    timeStyle: "short",
  });
}
