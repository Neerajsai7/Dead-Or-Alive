/**
 * charts.js — All Chart.js charts for LogiTwin dashboard
 */

let inventoryChart = null;
let vehicleChart   = null;
let onTimeChart    = null;
let invBarChart    = null;

const CHART_DEFAULTS = {
  color: "#e2e8f0",
  grid:  "#1e293b",
  font:  "Segoe UI"
};

// ── Inventory Bar Chart (Dashboard) ─────────────────────────
function updateInventoryChart(inventory) {
  const labels = inventory.map(i => i.name);
  const stock  = inventory.map(i => i.stock);
  const cap    = inventory.map(i => i.capacity);

  const colors = inventory.map(i =>
    i.status === "critical" ? "#dc2626" :
    i.status === "low"      ? "#f97316" : "#3b82f6"
  );

  const ctx = document.getElementById("inventoryChart");
  if (!ctx) return;

  if (inventoryChart) {
    inventoryChart.data.labels              = labels;
    inventoryChart.data.datasets[0].data   = stock;
    inventoryChart.data.datasets[0].backgroundColor = colors;
    inventoryChart.data.datasets[1].data   = cap;
    inventoryChart.update("none");
    return;
  }

  inventoryChart = new Chart(ctx, {
    type: "bar",
    data: {
      labels,
      datasets: [
        {
          label: "Stock",
          data: stock,
          backgroundColor: colors,
          borderRadius: 4,
        },
        {
          label: "Capacity",
          data: cap,
          backgroundColor: "#1e293b",
          borderRadius: 4,
        }
      ]
    },
    options: {
      responsive: true,
      animation: false,
      plugins: {
        legend: { labels: { color: CHART_DEFAULTS.color, font: { family: CHART_DEFAULTS.font } } }
      },
      scales: {
        x: { ticks: { color: CHART_DEFAULTS.color }, grid: { color: CHART_DEFAULTS.grid } },
        y: { ticks: { color: CHART_DEFAULTS.color }, grid: { color: CHART_DEFAULTS.grid } }
      }
    }
  });
}

// ── Vehicle Status Donut Chart ───────────────────────────────
function updateVehicleChart(vehicles) {
  const counts = { "in-transit": 0, delayed: 0, idle: 0 };
  vehicles.forEach(v => { if (counts[v.status] !== undefined) counts[v.status]++; });

  const ctx = document.getElementById("vehicleChart");
  if (!ctx) return;

  if (vehicleChart) {
    vehicleChart.data.datasets[0].data = Object.values(counts);
    vehicleChart.update("none");
    return;
  }

  vehicleChart = new Chart(ctx, {
    type: "doughnut",
    data: {
      labels: ["In Transit", "Delayed", "Idle"],
      datasets: [{
        data: Object.values(counts),
        backgroundColor: ["#3b82f6", "#f97316", "#64748b"],
        borderWidth: 0,
        hoverOffset: 6
      }]
    },
    options: {
      responsive: true,
      animation: false,
      plugins: {
        legend: { labels: { color: CHART_DEFAULTS.color, font: { family: CHART_DEFAULTS.font } } }
      }
    }
  });
}

// ── On-Time Rate Line Chart (Live History) ───────────────────
function updateOnTimeChart(history) {
  const labels = history.map((_, i) => i + 1);

  const ctx = document.getElementById("onTimeChart");
  if (!ctx) return;

  if (onTimeChart) {
    onTimeChart.data.labels            = labels;
    onTimeChart.data.datasets[0].data  = history;
    onTimeChart.update("none");
    return;
  }

  onTimeChart = new Chart(ctx, {
    type: "line",
    data: {
      labels,
      datasets: [{
        label: "On-Time Rate %",
        data: history,
        borderColor: "#3b82f6",
        backgroundColor: "rgba(59,130,246,0.1)",
        borderWidth: 2,
        tension: 0.4,
        fill: true,
        pointRadius: 2
      }]
    },
    options: {
      responsive: true,
      animation: false,
      plugins: {
        legend: { labels: { color: CHART_DEFAULTS.color } }
      },
      scales: {
        x: { ticks: { color: CHART_DEFAULTS.color }, grid: { color: CHART_DEFAULTS.grid } },
        y: {
          min: 0, max: 100,
          ticks: { color: CHART_DEFAULTS.color },
          grid:  { color: CHART_DEFAULTS.grid }
        }
      }
    }
  });
}

// ── Inventory Bar Chart (Inventory Page) ─────────────────────
function updateInvBarChart(inventory) {
  const ctx = document.getElementById("invBarChart");
  if (!ctx) return;

  const labels   = inventory.map(i => i.name);
  const stockPct = inventory.map(i => parseFloat(i.pct.toFixed(1)));
  const colors   = inventory.map(i =>
    i.status === "critical" ? "#dc2626" :
    i.status === "low"      ? "#f97316" : "#3b82f6"
  );

  if (invBarChart) {
    invBarChart.data.labels                        = labels;
    invBarChart.data.datasets[0].data              = stockPct;
    invBarChart.data.datasets[0].backgroundColor   = colors;
    invBarChart.update("none");
    return;
  }

  invBarChart = new Chart(ctx, {
    type: "bar",
    data: {
      labels,
      datasets: [{
        label: "Fill %",
        data: stockPct,
        backgroundColor: colors,
        borderRadius: 6
      }]
    },
    options: {
      responsive: true,
      animation: false,
      plugins: {
        legend: { labels: { color: CHART_DEFAULTS.color } }
      },
      scales: {
        x: { ticks: { color: CHART_DEFAULTS.color }, grid: { color: CHART_DEFAULTS.grid } },
        y: {
          min: 0, max: 100,
          ticks: { color: CHART_DEFAULTS.color, callback: v => v + "%" },
          grid:  { color: CHART_DEFAULTS.grid }
        }
      }
    }
  });
}