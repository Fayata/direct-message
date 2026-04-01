function parseJSONData(attr) {
  try {
    return JSON.parse(attr || "[]");
  } catch (_) {
    return [];
  }
}

function bindBulkSelection() {
  const selectAll = document.querySelector("[data-select-all]");
  if (!selectAll) return;
  selectAll.addEventListener("change", () => {
    document.querySelectorAll('input[name="ids"]').forEach((cb) => {
      cb.checked = selectAll.checked;
    });
  });
}

function bindConfirmActions() {
  document.querySelectorAll("form[data-confirm], button[data-confirm]").forEach((node) => {
    if (node.tagName === "FORM") {
      node.addEventListener("submit", (e) => {
        const msg = node.getAttribute("data-confirm") || "Lanjutkan?";
        if (!window.confirm(msg)) e.preventDefault();
      });
      return;
    }
    node.addEventListener("click", (e) => {
      const msg = node.getAttribute("data-confirm") || "Lanjutkan?";
      if (!window.confirm(msg)) e.preventDefault();
    });
  });
}

function drawTrendChart() {
  const canvas = document.getElementById("trendChart");
  if (!canvas) return;
  const labels = parseJSONData(canvas.dataset.labels);
  const peduli = parseJSONData(canvas.dataset.peduli);
  const gold = parseJSONData(canvas.dataset.gold);
  const ctx = canvas.getContext("2d");
  if (!ctx || labels.length === 0) return;

  const dpr = window.devicePixelRatio || 1;
  const width = canvas.clientWidth || 800;
  const height = canvas.clientHeight || 220;
  canvas.width = Math.floor(width * dpr);
  canvas.height = Math.floor(height * dpr);
  ctx.scale(dpr, dpr);

  const pad = { top: 16, right: 24, bottom: 28, left: 28 };
  const w = width - pad.left - pad.right;
  const h = height - pad.top - pad.bottom;
  const maxVal = Math.max(1, ...peduli, ...gold);

  function x(i) {
    if (labels.length === 1) return pad.left + w / 2;
    return pad.left + (i * w) / (labels.length - 1);
  }
  function y(v) {
    return pad.top + h - (v / maxVal) * h;
  }

  ctx.clearRect(0, 0, width, height);
  ctx.strokeStyle = "rgba(148,163,184,0.22)";
  ctx.lineWidth = 1;
  for (let i = 0; i <= 4; i += 1) {
    const gy = pad.top + (i * h) / 4;
    ctx.beginPath();
    ctx.moveTo(pad.left, gy);
    ctx.lineTo(width - pad.right, gy);
    ctx.stroke();
  }

  function drawSeries(values, color) {
    ctx.strokeStyle = color;
    ctx.lineWidth = 2;
    ctx.beginPath();
    values.forEach((v, i) => {
      const px = x(i);
      const py = y(v);
      if (i === 0) ctx.moveTo(px, py);
      else ctx.lineTo(px, py);
    });
    ctx.stroke();
    values.forEach((v, i) => {
      const px = x(i);
      const py = y(v);
      ctx.fillStyle = color;
      ctx.beginPath();
      ctx.arc(px, py, 2.8, 0, Math.PI * 2);
      ctx.fill();
    });
  }

  drawSeries(peduli, "#00e5a0");
  drawSeries(gold, "#d4af37");

  ctx.fillStyle = "rgba(156,163,175,0.9)";
  ctx.font = "11px system-ui, sans-serif";
  labels.forEach((lb, i) => {
    const px = x(i);
    ctx.textAlign = "center";
    ctx.fillText(lb, px, height - 8);
  });
}

function initAdminDashboard() {
  bindBulkSelection();
  bindConfirmActions();
  drawTrendChart();
  window.addEventListener("resize", drawTrendChart);
}

if (document.readyState === "loading") {
  document.addEventListener("DOMContentLoaded", initAdminDashboard);
} else {
  initAdminDashboard();
}
