document.addEventListener("DOMContentLoaded", function(event) {
// Ensure Chart.js is loaded
if (typeof Chart === 'undefined') {
  console.error('Chart.js is not loaded. Please ensure the library is included in your HTML.');
  return;
}

// Ensure the canvas element exists
const canvas = document.getElementById('halfyear-chart');
if (!canvas) {
  console.error('Canvas element not found');
  return;
}

// Parse the JSON data from the pre element
const pre = document.getElementById('halfyear-json');
if (!pre) {
  console.error('Pre element with id "halfyear-json" not found');
  return;
}

let halfYearTotals;
try {
  const jsonText = pre.textContent.trim();
  if (!jsonText || jsonText.startsWith('{{')) {
    console.error('Half-year totals JSON is missing or not rendered.');
    return;
  }
  halfYearTotals = JSON.parse(jsonText);
} catch (e) {
  console.error('Failed to parse half-year totals JSON:', e);
  return;
}

console.log(halfYearTotals);

const labels = halfYearTotals.map(item => item.Period);
const data = {
  Average: halfYearTotals.map(item => item.Average),
  LastTotal: halfYearTotals.map(item => item.LastTotal)
};

const ctx = canvas.getContext('2d');
if (!ctx) {
  console.error('Canvas context not found');
  return;
} else {
  console.log('Canvas context successfully retrieved');
}

new Chart(ctx, {
  type: 'bar',
  data: {
    labels: labels,
    datasets: [
      {
        label: 'Average',
        data: data.Average,
        borderWidth: 1
      },
      {
        label: 'Last Total',
        data: data.LastTotal,
        borderWidth: 1
      }
    ]
  },
  options: {
    scales: {
      y: {
        beginAtZero: true
      }
    }
  }
});
});
