function requireAuth() {
  if (!getAccessToken()) {
    window.location.href = "/login";
  }
}

function bindLogoutButton() {
  const btn = document.getElementById("logoutBtn");
  if (!btn) {
    return;
  }

  btn.addEventListener("click", async () => {
    await logout();
  });
}
