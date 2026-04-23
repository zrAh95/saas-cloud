async function login(email, password) {
  const response = await fetch("/api/v1/auth/login", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ email, password }),
  });

  const data = await response.json();
  if (!response.ok) {
    throw new Error(data.error || "Login gagal");
  }

  setTokens(data.data.access_token, data.data.refresh_token);
  window.location.href = "/dashboard";
}

async function logout() {
  const token = getAccessToken();
  if (token) {
    try {
      await fetch("/api/v1/auth/logout", {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });
    } catch (err) {
      console.error(err);
    }
  }

  clearTokens();
  window.location.href = "/login";
}
