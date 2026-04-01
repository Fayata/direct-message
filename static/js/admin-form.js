function initAdminFormPage() {
  document.querySelectorAll("form[data-confirm]").forEach((form) => {
    form.addEventListener("submit", (e) => {
      const message = form.getAttribute("data-confirm") || "Lanjutkan?";
      if (!window.confirm(message)) e.preventDefault();
    });
  });
}

if (document.readyState === "loading") {
  document.addEventListener("DOMContentLoaded", initAdminFormPage);
} else {
  initAdminFormPage();
}
