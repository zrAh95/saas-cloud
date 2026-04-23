let uploadTrafficChart = null;
let diskUsageChart = null;

async function loadDashboardStats() {
  try {
    const response = await apiFetch("/dashboard/stats");
    const stats = response.data || {};

    updateDashboardNumbers(stats);
    renderDashboardCharts(stats);
  } catch (error) {
    await showError(error.message, "Dashboard gagal dimuat");
  }
}

function updateDashboardNumbers(stats) {
  setText("totalFiles", pickField(stats, "total_files") ?? 0);
  setText("totalFolders", pickField(stats, "total_folders") ?? 0);
  setText("pendingShares", pickField(stats, "pending_shares") ?? 0);
  setText("usedStorage", formatGigabytes(pickField(stats, "used_storage") ?? 0));
  setText("freeStorage", formatGigabytes(pickField(stats, "free_storage") ?? 0));
  setText("totalDisk", formatGigabytes(pickField(stats, "total_disk") ?? 0));
  setText("diskUsagePct", `${Number(pickField(stats, "disk_usage_pct") || 0).toFixed(1)}%`);

  const diskProgress = document.getElementById("diskProgress");
  if (diskProgress) {
    diskProgress.style.width = `${Math.min(Number(pickField(stats, "disk_usage_pct") || 0), 100)}%`;
  }
}

function renderDashboardCharts(stats) {
  if (typeof ApexCharts === "undefined") {
    return;
  }

  const traffic = pickField(stats, "upload_traffic_7") || [];
  const labels = traffic.map((point) => point.label);
  const values = traffic.map((point) => Number(point.value || 0) / (1024 * 1024));
  const diskUsagePct = Number(pickField(stats, "disk_usage_pct") || 0);

  if (uploadTrafficChart) {
    uploadTrafficChart.updateOptions({ xaxis: { categories: labels } });
    uploadTrafficChart.updateSeries([{ name: "Upload MB", data: values }]);
  } else {
    uploadTrafficChart = new ApexCharts(document.querySelector("#uploadTrafficChart"), {
      chart: {
        type: "area",
        height: 300,
        toolbar: { show: false },
        foreColor: "#c2c7ff",
      },
      series: [{ name: "Upload MB", data: values }],
      xaxis: { categories: labels },
      stroke: { curve: "smooth", width: 3 },
      colors: ["#57caeb"],
      fill: {
        type: "gradient",
        gradient: { opacityFrom: 0.45, opacityTo: 0.05 },
      },
      dataLabels: { enabled: false },
      grid: { borderColor: "rgba(255,255,255,.08)" },
      yaxis: {
        labels: {
          formatter: (value) => `${value.toFixed(1)} MB`,
        },
      },
    });
    uploadTrafficChart.render();
  }

  if (diskUsageChart) {
    diskUsageChart.updateSeries([diskUsagePct]);
  } else {
    diskUsageChart = new ApexCharts(document.querySelector("#diskUsageChart"), {
      chart: {
        type: "radialBar",
        height: 280,
        foreColor: "#f2f7ff",
      },
      series: [diskUsagePct],
      colors: ["#5ddab4"],
      plotOptions: {
        radialBar: {
          hollow: { size: "68%" },
          dataLabels: {
            name: { show: true, text: "Disk Used" },
            value: {
              formatter: (value) => `${Number(value).toFixed(1)}%`,
            },
          },
        },
      },
      labels: ["Disk Used"],
    });
    diskUsageChart.render();
  }
}

function setText(id, value) {
  const element = document.getElementById(id);
  if (element) {
    element.textContent = value;
  }
}
